// PebbleDB implementation for DB interface.
// Provides efficient key-value storage for blocks, transactions, and other data.
// Supports batch operations and is optimized for concurrent access.
//
// Usage:
//   db, err := NewPebbledb(cfg)
//   err = db.Put(key, value)
//   value, err := db.Get(key)
//   err = db.Del(key)
//   err = db.Close()
//
// Batch operations:
//   err = db.BatchPut(key, value)   // enqueue
//   err = db.BatchPut(nil, nil)     // flush
//   err = db.BatchDel(key)          // enqueue delete
//   err = db.BatchDel(nil)          // flush

package pebbledb

import (
	"context"
	"errors"

	"github.com/cockroachdb/pebble"
	"github.com/rawbytedev/zerokv"
)

// pebbledb manages Database Insert/Deletion/Batch Operations for Pebble.
// It only handles []byte keys and values.
type pebbledb struct {
	db *pebble.DB
}
type pebbleBatch struct {
	batch *pebble.Batch
}

// NewPebbledb creates a new PebbleDB instance with the given config.
// Returns a StorageDB interface or an error if initialization fails.
func NewPebbledb(cfg Config) (zerokv.Core, error) {
	opts := &pebble.Options{}
	if cfg.PebbleConfigs != nil {
		opts = cfg.PebbleConfigs
	} else {
		opts = &pebble.Options{}
	}
	db, err := pebble.Open(cfg.Dir, opts)
	if err != nil {
		return nil, err
	}
	return &pebbledb{db: db}, nil
}

// Put inserts or updates a key-value pair in the database.
func (p *pebbledb) Put(ctx context.Context, key []byte, data []byte) error {
	return p.db.Set(key, data, pebble.Sync)
}

// Get retrieves the value for a given key. Returns an error if not found.
func (p *pebbledb) Get(ctx context.Context, key []byte) ([]byte, error) {
	val, closer, err := p.db.Get(key)
	if err != nil {
		return nil, err
	}
	defer closer.Close()
	return val, nil
}

// Del deletes a key-value pair from the database.
func (p *pebbledb) Delete(ctx context.Context, key []byte) error {
	return p.db.Delete(key, pebble.Sync)
}

// Close closes the database and releases all resources.
func (p *pebbledb) Close() error {
	var errs []error
	if err := p.db.Close(); err != nil {
		errs = append(errs, err)
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

func (b *pebbledb) Batch() zerokv.Batch {
	return &pebbleBatch{batch: b.db.NewBatch()}
}

func (p *pebbleBatch) Put(key []byte, data []byte) error {
	return p.batch.Set(key, data, pebble.NoSync)
}

// BatchDel adds a delete operation to the current batch.
func (p *pebbleBatch) Delete(key []byte) error {
	return p.batch.Delete(key, pebble.NoSync)
}

// flushBatch flushes any pending batch operations.
func (p *pebbleBatch) Commits(ctx context.Context) error {
	return p.batch.Commit(pebble.Sync)
}

func (b *pebbledb) Scan(prefix []byte) zerokv.Iterator {
	// Placeholder for Scan operation implementation
	return nil
}
func (b *pebbledb) NewIterator() zerokv.Iterator {
	// Placeholder for Iterator implementation
	return nil
}
func NewReverseIterator() zerokv.Iterator {
	// Placeholder for Reverse Iterator implementation
	return nil
}
func NewPrefixIterator(prefix []byte) zerokv.Iterator {
	// Placeholder for Prefix Iterator implementation
	return nil
}
func NewReversePrefixIterator(prefix []byte) zerokv.Iterator {
	// Placeholder for Reverse Prefix Iterator implementation
	return nil
}
