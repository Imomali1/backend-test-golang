package cache

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	errs "backend-test-golang/pkg/errors"
)

const testCleanUpInterval = 60

func TestMemCache_SetAndGet(t *testing.T) {
	cache := New(testCleanUpInterval)
	defer cache.Close()

	t.Run("normal set and get string value", func(t *testing.T) {
		key := "test-key"
		value := "test-value"
		ttl := 1 * time.Second

		cache.Set(key, value, ttl)

		got, err := cache.Get(key)
		if err != nil {
			t.Fatalf("got unexpected error: %v", err)
		}

		if got != value {
			t.Errorf("got: %v, want: %v", got, value)
		}
	})

	t.Run("get non-existent key", func(t *testing.T) {
		_, err := cache.Get("non-existent")
		if !errors.Is(err, errs.ErrNotFound) {
			t.Errorf("got error = %v, want %v", err, errs.ErrNotFound)
		}
	})

	t.Run("set and get struct value", func(t *testing.T) {
		type Payload struct {
			Name  string
			Value int
		}

		key := "struct-key"
		value := Payload{Name: "test", Value: 123}
		ttl := 1 * time.Second

		cache.Set(key, value, ttl)

		got, err := cache.Get(key)
		if err != nil {
			t.Fatalf("got unexpected error: %v", err)
		}

		payload, ok := got.(Payload)
		if !ok {
			t.Fatalf("returned wrong type")
		}

		if payload.Name != value.Name || payload.Value != value.Value {
			t.Errorf("got %+v, want %+v", payload, value)
		}
	})
}

func TestMemCache_TTL(t *testing.T) {
	cache := New(testCleanUpInterval)
	defer cache.Close()

	t.Run("item expires after TTL", func(t *testing.T) {
		key := "expired"
		value := "to expire"
		ttl := 100 * time.Millisecond

		cache.Set(key, value, ttl)

		_, err := cache.Get(key)
		if err != nil {
			t.Errorf("Get method should succeed immediately after Set method call")
		}

		time.Sleep(150 * time.Millisecond)

		_, err = cache.Get(key)
		if !errors.Is(err, errs.ErrNotFound) {
			t.Errorf("Get method after TTL should return ErrNotFound, got %v", err)
		}
	})

	t.Run("item still valid before TTL", func(t *testing.T) {
		key := "valid test"
		value := "still valid"
		ttl := 500 * time.Millisecond

		cache.Set(key, value, ttl)

		time.Sleep(100 * time.Millisecond)

		got, err := cache.Get(key)
		if err != nil {
			t.Errorf("Get method should succeed before TTL expires: %v", err)
		}

		if got != value {
			t.Errorf("got %v, want %v", got, value)
		}
	})
}

func TestMemCache_Delete(t *testing.T) {
	cache := New(testCleanUpInterval)
	defer cache.Close()

	t.Run("delete existing key", func(t *testing.T) {
		key := "delete test"
		value := "will be deleted"
		ttl := 1 * time.Second

		cache.Set(key, value, ttl)

		_, err := cache.Get(key)
		if err != nil {
			t.Fatalf("Get method before delete failed: %v", err)
		}

		cache.Delete(key)

		_, err = cache.Get(key)
		if !errors.Is(err, errs.ErrNotFound) {
			t.Errorf("Get method after delete should return ErrNotFound, got %v", err)
		}
	})

	t.Run("delete non-existent key does not error", func(t *testing.T) {
		cache.Delete("non-existent-key")
	})

	t.Run("delete multiple keys", func(t *testing.T) {
		ttl := 1 * time.Second
		cache.Set("key1", "value1", ttl)
		cache.Set("key2", "value2", ttl)
		cache.Set("key3", "value3", ttl)

		cache.Delete("key1", "key2")

		_, err := cache.Get("key1")
		if !errors.Is(err, errs.ErrNotFound) {
			t.Error("key1 should be deleted")
		}

		_, err = cache.Get("key2")
		if !errors.Is(err, errs.ErrNotFound) {
			t.Error("key2 should be deleted")
		}

		_, err = cache.Get("key3")
		if err != nil {
			t.Error("key3 should still exist")
		}
	})
}

func TestMemCache_Concurrency(t *testing.T) {
	cache := New(testCleanUpInterval)
	defer cache.Close()

	t.Run("concurrent reads and writes", func(t *testing.T) {
		const goroutines = 100
		const operations = 100

		var wg sync.WaitGroup
		wg.Add(goroutines * 2)

		for i := 0; i < goroutines; i++ {
			go func(id int) {
				defer wg.Done()
				for j := 0; j < operations; j++ {
					key := "concurrent-key"
					value := id*operations + j
					cache.Set(key, value, 1*time.Second)
				}
			}(i)
		}

		for range goroutines {
			go func() {
				defer wg.Done()
				for range operations {
					cache.Get("concurrent-key")
				}
			}()
		}

		wg.Wait()
	})

	t.Run("concurrent writes to different keys", func(t *testing.T) {
		const goroutines = 50

		var wg sync.WaitGroup
		wg.Add(goroutines)

		for i := 0; i < goroutines; i++ {
			go func(id int) {
				defer wg.Done()
				key := "key-" + string(rune(id))
				cache.Set(key, id, 1*time.Second)
			}(i)
		}

		wg.Wait()

		for i := 0; i < goroutines; i++ {
			key := "key-" + string(rune(i))
			val, err := cache.Get(key)
			if err != nil {
				t.Errorf("key %s should exist", key)
				continue
			}
			if val != i {
				t.Errorf("key: %s, got: %v, want: %v", key, val, i)
			}
		}
	})
}

func TestMemCache_Cleanup(t *testing.T) {
	cache := New(1) // 1 second to test clean up
	defer cache.Close()

	t.Run("expired items are cleaned up", func(t *testing.T) {
		ttl := 50 * time.Millisecond

		for i := 0; i < 10; i++ {
			key := fmt.Sprintf("to-cleanup-%d", i)
			cache.Set(key, i, ttl)
		}

		if cache.Len() != 10 {
			t.Errorf("cache.Len() should be 10, got %v", cache.Len())
		}

		time.Sleep(200 * time.Millisecond)

		for i := 0; i < 10; i++ {
			key := "cleanup-" + string(rune(i))
			_, err := cache.Get(key)
			if !errors.Is(err, errs.ErrNotFound) {
				t.Errorf("key %s should be expired", key)
			}
		}

		if cache.Len() != 10 {
			t.Errorf("cache.Len() should be 10, got %v", cache.Len())
		}

		time.Sleep(time.Second)

		if cache.Len() != 0 {
			t.Errorf("cache.Len() should be 0, got %v", cache.Len())
		}
	})
}

func TestMemCache_Close(t *testing.T) {
	t.Run("close stops cleanup goroutine", func(t *testing.T) {
		cache := New(testCleanUpInterval)

		cache.Set("test", "value", 1*time.Second)

		err := cache.Close()
		if err != nil {
			t.Errorf("got unexpected error: %v", err)
		}

		err = cache.Close()
		if err != nil {
			t.Errorf("Second Close method got unexpected error: %v", err)
		}
	})

	t.Run("operations after close", func(t *testing.T) {
		cache := New(testCleanUpInterval)
		cache.Set("test", "value", 1*time.Second)
		cache.Close()

		_, err := cache.Get("test")
		if !errors.Is(err, errs.ErrNotFound) {
			t.Errorf("Get after Close should return ErrNotFound, got %v", err)
		}
	})
}

func BenchmarkMemCache_Set(b *testing.B) {
	cache := New(testCleanUpInterval)
	defer cache.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set("bench-key", i, 1*time.Second)
	}
}

func BenchmarkMemCache_Get(b *testing.B) {
	cache := New(testCleanUpInterval)
	defer cache.Close()

	cache.Set("bench-key", "value", 1*time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get("bench-key")
	}
}

func BenchmarkMemCache_ConcurrentAccess(b *testing.B) {
	cache := New(testCleanUpInterval)
	defer cache.Close()

	cache.Set("bench-key", "value", 1*time.Second)

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%2 == 0 {
				cache.Get("bench-key")
			} else {
				cache.Set("bench-key", i, 1*time.Second)
			}
			i++
		}
	})
}
