package cache

import (
	"log"
	"sync"
	"time"

	errs "backend-test-golang/pkg/errors"
)

type MemCache struct {
	mem      map[string]cacheItem
	mu       sync.RWMutex
	wg       sync.WaitGroup
	syncOnce sync.Once
	stopCh   chan struct{}
}

type cacheItem struct {
	payload   any
	expiresAt time.Time
}

func New(cacheCleanUpIntervalSeconds int) *MemCache {
	m := &MemCache{
		mem:    make(map[string]cacheItem),
		stopCh: make(chan struct{}),
	}
	m.wg.Add(1)
	go m.cleanUp(time.Duration(cacheCleanUpIntervalSeconds) * time.Second)
	return m
}

func (c *MemCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.mem)
}

func (c *MemCache) Get(key string) (any, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.mem[key]
	if !found {
		return nil, errs.ErrNotFound
	}

	if time.Now().After(item.expiresAt) {
		return nil, errs.ErrNotFound
	}

	return item.payload, nil
}

func (c *MemCache) Set(key string, payload any, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.mem[key] = cacheItem{
		payload:   payload,
		expiresAt: time.Now().Add(ttl),
	}
}

func (c *MemCache) Delete(keys ...string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, key := range keys {
		delete(c.mem, key)
	}
}

func (c *MemCache) Close() error {
	c.syncOnce.Do(func() {
		close(c.stopCh)
		c.wg.Wait()
		c.mu.Lock()
		c.mem = nil
		c.mu.Unlock()
		log.Println("mem cache closed and cleaned")
	})

	return nil
}

func (c *MemCache) cleanUp(interval time.Duration) {
	defer c.wg.Done()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println("mem cache cleaning up")
			c.evictExpired()
		case <-c.stopCh:
			return
		}
	}
}

func (c *MemCache) evictExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, item := range c.mem {
		if now.After(item.expiresAt) {
			delete(c.mem, key)
		}
	}
}
