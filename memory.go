package otp

import "time"

type Entry struct {
	content *string
	views   int
}

type MemoryStore map[string]Entry

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

func (ms *MemoryStore) Get(key string) *string {
	result := (*ms)[key]

	result.views = result.views - 1
	(*ms)[key] = result

	if result.views == 0 {
		if _, ok := (*ms)[key]; ok {
			delete(*ms, key)
		}
	}

	return result.content
}

func (ms *MemoryStore) Exists(key string) bool {
	_, ok := (*ms)[key]
	return ok
}

func (ms *MemoryStore) Set(content string, views int, expire int) string {
	key := generateUUID()
	(*ms)[key] = Entry{
		content: &content,
		views:   views,
	}

	go func() {
		time.Sleep(time.Second * time.Duration(expire))
		if _, ok := (*ms)[key]; ok {
			delete(*ms, key)
		}
	}()

	return key
}
