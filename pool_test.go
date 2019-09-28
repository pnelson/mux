package mux

import (
	"bytes"
	"testing"
)

func TestPool(t *testing.T) {
	b := &pool{free: make(chan *bytes.Buffer, 1)}
	buf := b.Get()
	if len(b.free) != 0 {
		t.Fatalf("should create new buffer")
	}
	b.Put(buf)
	if len(b.free) != 1 {
		t.Fatalf("should add buffer to free list")
	}
	b.Put(bytes.NewBuffer(make([]byte, 0)))
	if len(b.free) != 1 {
		t.Fatalf("should discard buffer if free list is full")
	}
	b.Get()
	if len(b.free) != 0 {
		t.Fatalf("should reuse buffer from free list")
	}
}
