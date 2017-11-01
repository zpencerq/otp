package otp

import "time"

type Entry struct {
	content *string
	views   int
}

type MemoryStore map[string]Entry

type MemoryConn struct {
	store *MemoryStore
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

func (ms *MemoryStore) NewConn() OneTimeStoreConn {
	return &MemoryConn{ms}
}

func (mc *MemoryConn) Get(key string) *string {
	result := (*mc.store)[key]

	result.views = result.views - 1
	(*mc.store)[key] = result

	if result.views == 0 {
		if _, ok := (*mc.store)[key]; ok {
			delete(*mc.store, key)
		}
	}

	return result.content
}

func (mc *MemoryConn) Exists(key string) bool {
	_, ok := (*mc.store)[key]
	return ok
}

func (mc *MemoryConn) Set(content string, views int, expire int) string {
	key := generateUUID()
	(*mc.store)[key] = Entry{
		content: &content,
		views:   views,
	}

	go func() {
		time.Sleep(time.Second * time.Duration(expire))
		if _, ok := (*mc.store)[key]; ok {
			delete(*mc.store, key)
		}
	}()

	return key
}

func (rc *MemoryConn) Close() error {
	return nil
}
