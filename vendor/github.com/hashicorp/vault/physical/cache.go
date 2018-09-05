package physical

import (
	"context"
	"sync/atomic"

	"github.com/hashicorp/golang-lru"
	"github.com/hashicorp/vault/helper/locksutil"
	log "github.com/mgutz/logxi/v1"
)

const (
	// DefaultCacheSize is used if no cache size is specified for NewCache
	DefaultCacheSize = 128 * 1024
)

// Cache is used to wrap an underlying physical backend
// and provide an LRU cache layer on top. Most of the reads done by
// Vault are for policy objects so there is a large read reduction
// by using a simple write-through cache.
type Cache struct {
	backend Backend
	lru     *lru.TwoQueueCache
	locks   []*locksutil.LockEntry
	logger  log.Logger
	enabled *uint32
}

// TransactionalCache is a Cache that wraps the physical that is transactional
type TransactionalCache struct {
	*Cache
	Transactional
}

// Verify Cache satisfies the correct interfaces
var _ ToggleablePurgemonster = (*Cache)(nil)
var _ ToggleablePurgemonster = (*TransactionalCache)(nil)
var _ Backend = (*Cache)(nil)
var _ Transactional = (*TransactionalCache)(nil)

// NewCache returns a physical cache of the given size.
// If no size is provided, the default size is used.
func NewCache(b Backend, size int, logger log.Logger) *Cache {
	if logger.IsTrace() {
		logger.Trace("physical/cache: creating LRU cache", "size", size)
	}
	if size <= 0 {
		size = DefaultCacheSize
	}

	cache, _ := lru.New2Q(size)
	c := &Cache{
		backend: b,
		lru:     cache,
		locks:   locksutil.CreateLocks(),
		logger:  logger,
		// This fails safe.
		enabled: new(uint32),
	}
	return c
}

func NewTransactionalCache(b Backend, size int, logger log.Logger) *TransactionalCache {
	c := &TransactionalCache{
		Cache:         NewCache(b, size, logger),
		Transactional: b.(Transactional),
	}
	return c
}

// SetEnabled is used to toggle whether the cache is on or off. It must be
// called with true to actually activate the cache after creation.
func (c *Cache) SetEnabled(enabled bool) {
	if enabled {
		atomic.StoreUint32(c.enabled, 1)
		return
	}
	atomic.StoreUint32(c.enabled, 0)
}

// Purge is used to clear the cache
func (c *Cache) Purge(ctx context.Context) {
	// Lock the world
	for _, lock := range c.locks {
		lock.Lock()
		defer lock.Unlock()
	}

	c.lru.Purge()
}

func (c *Cache) Put(ctx context.Context, entry *Entry) error {
	if atomic.LoadUint32(c.enabled) == 0 {
		return c.backend.Put(ctx, entry)
	}

	lock := locksutil.LockForKey(c.locks, entry.Key)
	lock.Lock()
	defer lock.Unlock()

	err := c.backend.Put(ctx, entry)
	if err == nil {
		c.lru.Add(entry.Key, entry)
	}
	return err
}

func (c *Cache) Get(ctx context.Context, key string) (*Entry, error) {
	if atomic.LoadUint32(c.enabled) == 0 {
		return c.backend.Get(ctx, key)
	}

	lock := locksutil.LockForKey(c.locks, key)
	lock.RLock()
	defer lock.RUnlock()

	// Check the LRU first
	if raw, ok := c.lru.Get(key); ok {
		if raw == nil {
			return nil, nil
		}
		return raw.(*Entry), nil
	}

	// Read from the underlying backend
	ent, err := c.backend.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if ent != nil {
		c.lru.Add(key, ent)
	}

	return ent, nil
}

func (c *Cache) Delete(ctx context.Context, key string) error {
	if atomic.LoadUint32(c.enabled) == 0 {
		return c.backend.Delete(ctx, key)
	}

	lock := locksutil.LockForKey(c.locks, key)
	lock.Lock()
	defer lock.Unlock()

	err := c.backend.Delete(ctx, key)
	if err == nil {
		c.lru.Remove(key)
	}
	return err
}

func (c *Cache) List(ctx context.Context, prefix string) ([]string, error) {
	// Always pass-through as this would be difficult to cache. For the same
	// reason we don't lock as we can't reasonably know which locks to readlock
	// ahead of time.
	return c.backend.List(ctx, prefix)
}

func (c *TransactionalCache) Transaction(ctx context.Context, txns []*TxnEntry) error {
	// Collect keys that need to be locked
	var keys []string
	for _, curr := range txns {
		keys = append(keys, curr.Entry.Key)
	}
	// Lock the keys
	for _, l := range locksutil.LocksForKeys(c.locks, keys) {
		l.Lock()
		defer l.Unlock()
	}

	if err := c.Transactional.Transaction(ctx, txns); err != nil {
		return err
	}

	if atomic.LoadUint32(c.enabled) == 1 {
		for _, txn := range txns {
			switch txn.Operation {
			case PutOperation:
				c.lru.Add(txn.Entry.Key, txn.Entry)
			case DeleteOperation:
				c.lru.Remove(txn.Entry.Key)
			}
		}
	}

	return nil
}
