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
		{"https://example.com", []byte("testdata")},
		{"https://example.com/path", []byte("moredata")},
	}

	for i, cse := range cases {
		t.Run(fmt.Sprintf("case-%d", i), func(t *testing.T) {
			c := NewCache(interval)
			c.Add(cse.key, cse.val)
			got, ok := c.Get(cse.key)
			if !ok {
				t.Fatalf("expected key")
			}
			if string(got) != string(cse.val) {
				t.Fatalf("value mismatch")
			}
		})
	}
}

func TestReapLoop(t *testing.T) {
	const iv = 5 * time.Millisecond
	cache := NewCache(iv)
	cache.Add("k", []byte("x"))

	if _, ok := cache.Get("k"); !ok {
		t.Fatalf("expected key")
	}
	time.Sleep(iv + 5*time.Millisecond) // wait past expiry

	if _, ok := cache.Get("k"); ok {
		t.Fatalf("expected key to be reaped")
	}
}
