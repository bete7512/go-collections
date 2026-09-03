package main

import (
	"sync"
	"sync/atomic"
)

func main() {}

// unsafe
type UnsafeCounter struct {
	count int
}

func (c *UnsafeCounter) Inc() {
	c.count++
}
func (c *UnsafeCounter) Value() int {
	return c.count
}

// safe
type SafeCounter struct {
	mu sync.Mutex
	n  int
}

func (c *SafeCounter) Inc() {
	defer c.mu.Unlock()
	c.mu.Lock()
	c.n++
}
func (c *SafeCounter) Value() int {
	defer c.mu.Unlock()
	c.mu.Lock()
	return c.n
}

type AtomicCounter struct {
	count atomic.Int64
}

func (c *AtomicCounter) Inc(){
	c.count.Add(1)
}
func (c *AtomicCounter) Value() int64{
	return c.count.Load()
}
