package interfaces

type Repository interface {
	Get(key string) ([]byte, error)
	Set(key string, entry []byte) error
}
