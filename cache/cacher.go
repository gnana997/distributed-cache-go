package cache

import "time"

type Cacher interface {
	Has([]byte) bool
	Set([]byte, []byte, time.Duration) error
	Get([]byte) ([]byte, error)
	Delete([]byte) error
}
