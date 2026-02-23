package handler

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"imagesv2/internal/service"
)

type Handler struct {
	svc *service.Service
}

func New(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(r *gin.Engine) {
	r.GET("/health", h.Health)
	r.POST("/images", h.Upload)
	r.GET("/images", h.List)
	r.GET("/exists/*path", h.Exists)
	r.NoRoute(h.Serve)
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

type uploadRequest struct {
	URL string `json:"url" binding:"required"`
}

func (h *Handler) Upload(c *gin.Context) {
	var req uploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	publicURL, err := h.svc.UploadFromURL(c.Request.Context(), req.URL)
	if err != nil {
		status := classifyError(err)
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": publicURL})
}

func (h *Handler) List(c *gin.Context) {
	urls, err := h.svc.ListImages()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	images := make([]gin.H, len(urls))
	for i, u := range urls {
		images[i] = gin.H{"url": u}
	}
	c.JSON(http.StatusOK, gin.H{"images": images})
}

func (h *Handler) Exists(c *gin.Context) {
	path := c.Param("path")
	c.JSON(http.StatusOK, gin.H{"exists": h.svc.ImageExists(path)})
}

func (h *Handler) Serve(c *gin.Context) {
	if c.Request.Method != http.MethodGet {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
		return
	}

	path := c.Request.URL.Path
	data, err := h.svc.GetImage(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			c.JSON(http.StatusNotFound, gin.H{"error": "image not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	contentType := http.DetectContentType(data)
	c.Data(http.StatusOK, contentType, data)
}

func classifyError(err error) int {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "invalid url"),
		strings.Contains(msg, "unsupported scheme"),
		strings.Contains(msg, "no file path"):
		return http.StatusBadRequest
	case strings.Contains(msg, "not an image"):
		return http.StatusBadRequest
	case strings.Contains(msg, "file too large"):
		return http.StatusRequestEntityTooLarge
	case strings.Contains(msg, "download failed"),
		strings.Contains(msg, "unexpected status"):
		return http.StatusBadGateway
	default:
		return http.StatusInternalServerError
	}
}
