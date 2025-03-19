package zch

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

type Item struct {
	value      string
	expiration int64
}

func (item Item) Expired() bool {
	if item.expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > item.expiration
}

const (
	NoExpiration      time.Duration = -1
	DefaultExpiration time.Duration = 0
)

type Memory struct {
	*memory
}

type memory struct {
	defaultExpiration time.Duration
	items             map[string]Item
	mu                sync.RWMutex
	janitor           *janitor
}

func (c *memory) Set(k string, x string, d time.Duration) {
	var e int64
	if d == DefaultExpiration {
		d = c.defaultExpiration
	}
	if d > 0 {
		e = time.Now().Add(d).UnixNano()
	}
	c.mu.Lock()
	c.items[k] = Item{
		value:      x,
		expiration: e,
	}
	c.mu.Unlock()
}

func (c *memory) set(k string, x string, d time.Duration) {
	var e int64
	if d == DefaultExpiration {
		d = c.defaultExpiration
	}
	if d > 0 {
		e = time.Now().Add(d).UnixNano()
	}
	c.items[k] = Item{
		value:      x,
		expiration: e,
	}
}

func (c *memory) SetDefault(k string, x string) {
	c.Set(k, x, DefaultExpiration)
}

func (c *memory) SetNX(k string, x string, d time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, found := c.get(k)
	if found {
		return fmt.Errorf("item %s already exists", k)
	}
	c.set(k, x, d)
	return nil
}

func (c *memory) Replace(k string, x string, d time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, found := c.get(k)
	if !found {
		return fmt.Errorf("item %s doesn't exist", k)
	}
	c.set(k, x, d)
	return nil
}

func (c *memory) Get(k string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, found := c.items[k]
	if !found {
		return "", false
	}
	if item.expiration > 0 {
		if time.Now().UnixNano() > item.expiration {
			return "", false
		}
	}
	return item.value, true
}

func (c *memory) GetWithExpiration(k string) (string, time.Time, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, found := c.items[k]
	if !found {
		return "", time.Time{}, false
	}

	if item.expiration > 0 {
		if time.Now().UnixNano() > item.expiration {
			return "", time.Time{}, false
		}
		return item.value, time.Unix(0, item.expiration), true
	}
	return item.value, time.Time{}, true
}

func (c *memory) get(k string) (string, bool) {
	item, found := c.items[k]
	if !found {
		return "", false
	}
	if item.expiration > 0 {
		if time.Now().UnixNano() > item.expiration {
			return "", false
		}
	}
	return item.value, true
}

func (c *memory) Delete(k string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, k)
}

func (c *memory) DeleteExpired() {
	now := time.Now().UnixNano()
	for k, v := range c.items {
		if v.expiration > 0 && now > v.expiration {
			c.Delete(k)
		}
	}
}

func (c *memory) Items() map[string]Item {
	c.mu.RLock()
	defer c.mu.RUnlock()
	m := make(map[string]Item, len(c.items))
	now := time.Now().UnixNano()
	for k, v := range c.items {
		if v.expiration > 0 {
			if now > v.expiration {
				continue
			}
		}
		m[k] = v
	}
	return m
}

func (c *memory) ItemCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	n := len(c.items)
	return n
}

func (c *memory) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = map[string]Item{}
}

type janitor struct {
	Interval time.Duration
	stop     chan bool
}

func (j *janitor) Run(c *memory) {
	ticker := time.NewTicker(j.Interval)
	for {
		select {
		case <-ticker.C:
			c.DeleteExpired()
		case <-j.stop:
			ticker.Stop()
			return
		}
	}
}

func stopJanitor(c *Memory) {
	c.janitor.stop <- true
}

func runJanitor(c *memory, ci time.Duration) {
	j := &janitor{
		Interval: ci,
		stop:     make(chan bool),
	}
	c.janitor = j
	go j.Run(c)
}

func newMemory(de time.Duration, m map[string]Item) *memory {
	if de == 0 {
		de = -1
	}
	c := &memory{
		defaultExpiration: de,
		items:             m,
	}
	return c
}

func newMemoryWithJanitor(de time.Duration, ci time.Duration, m map[string]Item) *Memory {
	c := newMemory(de, m)
	C := &Memory{c}
	if ci > 0 {
		runJanitor(c, ci)
		runtime.SetFinalizer(C, stopJanitor)
	}
	return C
}

func NewMemory(defaultExpiration, cleanupInterval time.Duration) *Memory {
	items := make(map[string]Item)
	return newMemoryWithJanitor(defaultExpiration, cleanupInterval, items)
}

func NewMemoryFrom(defaultExpiration, cleanupInterval time.Duration, items map[string]Item) *Memory {
	return newMemoryWithJanitor(defaultExpiration, cleanupInterval, items)
}
