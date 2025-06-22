package interfaces

import (
	"time"
)

type Repository interface {
	Get(key string) ([]byte, error)
	Set(key string, entry []byte) error
	GetWithExpiration(key string) ([]byte, error)
	SetWithExpiration(key string, entry []byte, ttl time.Duration) error
}
