package mux

import "bytes"

// Pool represents a buffer pool.
type Pool interface {
	Get() *bytes.Buffer
	Put(b *bytes.Buffer)
}

// pool is a naive bounded buffer pool implementation.
type pool struct {
	free chan *bytes.Buffer
}

// Get retrieves a buffer from the pool if available
// or allocates a new one if not.
func (p *pool) Get() *bytes.Buffer {
	select {
	case b := <-p.free:
		return b
	default:
		return bytes.NewBuffer(make([]byte, 0))
	}
}

// Put adds b to the free list unless the list is full,
// in which case the buffer is discarded.
func (p *pool) Put(b *bytes.Buffer) {
	b.Reset()
	select {
	case p.free <- b:
	default:
	}
}
