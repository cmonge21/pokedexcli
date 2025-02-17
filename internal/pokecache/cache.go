package pokecache

import (
	"sync"
)

tyoe Cache struct {
	createdAt time.Time
	val []byte
}

func NewCach() *Cache {
	return &Cache{
		data: make(map[string]cacheEntry),
	}
}