package storage

type Storage interface {
	Save(path string, data []byte) error
	Get(path string) ([]byte, error)
	Exists(path string) bool
	List() ([]string, error)
}
