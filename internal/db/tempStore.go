package db

import (
	"strings"
	"sync"
)

var tempStore sync.Map

func SaveTemporary(key, value string) error {
	tempStore.Store(key, value)
	return nil
}

func LoadTemporary(key string) (string, bool) {
	v, ok := tempStore.Load(key)
	if !ok {
		return "", false
	}
	return v.(string), true
}

func ClearTemporaryByPrefix(prefix string) {
	tempStore.Range(func(k, _ any) bool {
		if key, ok := k.(string); ok && strings.HasPrefix(key, prefix) {
			tempStore.Delete(key)
		}
		return true
	})
}
