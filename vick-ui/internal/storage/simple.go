package storage

import (
	"sync"
	"maps"
)

type SimpleStorage struct {
	data map[string]any
	mutex sync.RWMutex
}

func NewSimpleStorage() *SimpleStorage {
	return &SimpleStorage{
		data: make(map[string]any),
	}
}

func (db *SimpleStorage) GetData() map[string]any {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	result := make(map[string]any)
	maps.Copy(result, db.data)

	return result
}

func (db *SimpleStorage) Insert(key string, value any) {
	db.mutex.Lock()
	db.data[key] = value
	db.mutex.Unlock()
}

func (db *SimpleStorage) Remove(key string) {
	db.mutex.Lock()
	delete(db.data, key)
	db.mutex.Unlock()
}

func (db *SimpleStorage) Search(key string) (any, bool) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	if v, ok := db.data[key]; ok {
		return v, ok
	}

	return nil, false
}

