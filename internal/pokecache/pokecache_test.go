package pokecache

import (
	"fmt"
	"testing"
	"time"
)

func TestAddGet(t *testing.T) {
	const interval = 5 * time.Second
	cases := []struct {
		key string
		val []byte
	}{
		{
			key: "https://example.com",
			val: []byte("testdata"),
		},
		{
			key: "https://example.com/path",
			val: []byte("moretestdata"),
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("Test case %v", i), func(t *testing.T) {
			cache := NewCache(interval)
			cache.Add(c.key, c.val)
			val, ok := cache.Get(c.key)
			if !ok {
				t.Errorf("expected to find key")
				return
			}
			if string(val) != string(c.val) {
				t.Errorf("expected to find value")
				return
			}
		})
	}
}

func TestGetMissing(t *testing.T) {
	cache := NewCache(5 * time.Second)
	if _, ok := cache.Get("https://example.com/never-added"); ok {
		t.Errorf("expected missing key to report ok == false")
	}
}

func TestAddOverwrite(t *testing.T) {
	cache := NewCache(5 * time.Second)
	const key = "https://example.com"
	cache.Add(key, []byte("first"))
	cache.Add(key, []byte("second"))

	val, ok := cache.Get(key)
	if !ok {
		t.Fatalf("expected to find key after overwrite")
	}
	if string(val) != "second" {
		t.Errorf("expected overwritten value %q, got %q", "second", string(val))
	}
}

func TestReapLoop(t *testing.T) {
	const baseTime = 5 * time.Millisecond
	const waitTime = baseTime + 5*time.Millisecond
	cache := NewCache(baseTime)
	cache.Add("https://example.com", []byte("testdata"))

	_, ok := cache.Get("https://example.com")
	if !ok {
		t.Errorf("expected to find key")
		return
	}

	time.Sleep(waitTime)

	_, ok = cache.Get("https://example.com")
	if ok {
		t.Errorf("expected to not find key")
		return
	}
}
