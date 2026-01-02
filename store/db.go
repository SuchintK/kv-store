package store

import (
	"time"
)

func Set(key string, value *Value) {
	mut.Lock()
	defer mut.Unlock()
	db[key] = value
}

func Get(key string) (*Value, bool) {
	mut.Lock()
	defer mut.Unlock()
	value, exist := db[key]
	if exist {
		if value.ExpiresAt != nil && value.ExpiresAt.Before(time.Now()) {
			delete(db, key)
			return &Value{}, false
		}
		return value, exist
	}
	return &Value{}, false
}
func Delete(key string) {
	mut.Lock()
	defer mut.Unlock()
	delete(db, key)
}
