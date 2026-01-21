package dbs

// StorageDB defines the interface for a pluggable key-value store.
type DB interface {
	// Put inserts or updates a key-value pair in the database.
	Put(key []byte, data []byte) error
	// Get retrieves the value for a given key. Returns an error if not found.
	Get(key []byte) ([]byte, error)
	// Del deletes a key-value pair from the database.
	Del(key []byte) error
	// BatchPut adds a key-value pair to the current batch. 
	BatchPut(key []byte, data []byte) error
	// BatchDel adds a delete operation to the current batch.
	BatchDel(key []byte) error
	 // FlushBatch flushes any pending batch operations.
    FlushBatch() error
	// Close closes the database and releases all resources.
	Close() error
}
