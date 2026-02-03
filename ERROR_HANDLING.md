# Error Handling Guide

ZeroKV implementations may handle errors differently depending on the underlying database. This guide documents the error behavior of each implementation and provides guidance for handling these differences.

## Important Principle

**ZeroKV does not control or alter how the underlying database returns errors.** If the underlying database panics, returns an error, or behaves in a specific way, ZeroKV passes that behavior through to the user.

When implementing a new backend, you must handle these database-specific behaviors and either:

1. Convert them to errors and return them consistently
2. Document the specific behavior clearly

## BadgerDB Error Behavior

BadgerDB consistently returns errors for error conditions.

### Batch Operations

**After Commit:**

- Attempting `Put()` after `Commit()` returns an error
- Attempting `Delete()` after `Commit()` returns an error
- Attempting another `Commit()` returns an error

```go
batch := db.Batch()
batch.Put([]byte("key1"), []byte("value1"))
batch.Commit(context.Background())

// Returns error: "This transaction has been discarded. Create a new one"
err := batch.Put([]byte("key2"), []byte("value2"))
if err != nil {
    log.Println("Error:", err)
}
```

### Iterator Behavior

- `Error()` returns `nil` if no errors occurred
- `Error()` returns the most recent error if one occurred
- Always safe to call even after iteration completes

```go
iterator := db.Scan([]byte("prefix"))
defer iterator.Release()

for iterator.Next() {
    // Process key/value
}

if iterator.Error() != nil {
    log.Fatal(iterator.Error())
}
```

### Key Not Found

`Get()` returns an error when a key is not found:

```go
value, err := db.Get(context.Background(), []byte("nonexistent"))
if err != nil {
    log.Println("Key not found or I/O error:", err)
}
```

## PebbleDB Error Behavior

PebbleDB has different error handling characteristics, particularly with batch operations.

### Batch Operations (PebbleDB)

**After Commit:**

- Attempting `Put()` after `Commit()` **panics**
- Attempting `Delete()` after `Commit()` **panics**
- Attempting another `Commit()` **panics**

```go
batch := db.Batch()
batch.Put([]byte("key1"), []byte("value1"))
batch.Commit(context.Background())

// This will panic!
batch.Put([]byte("key2"), []byte("value2"))
```

**Example with recovery:**

```go
defer func() {
    if r := recover(); r != nil {
        log.Println("Panic recovered:", r)
    }
}()

batch := db.Batch()
batch.Put([]byte("key"), []byte("value"))
batch.Commit(context.Background())
batch.Put([]byte("key2"), []byte("value2")) // Panics
```

### Iterator Behavior (PebbleDB)

- `Error()` may panic if called when no errors occurred (empty slice access)
- **FIXED:** Error handling now includes nil check to prevent panics

```go
iterator := db.Scan([]byte("prefix"))
defer iterator.Release()

for iterator.Next() {
    // Process key/value
}

// Safe to call - returns nil if no errors
if iterator.Error() != nil {
    log.Fatal(iterator.Error())
}
```

### Key Not Found (PebbleDB)

`Get()` returns an error when a key is not found (similar to BadgerDB):

```go
value, err := db.Get(context.Background(), []byte("nonexistent"))
if err != nil {
    log.Println("Key not found or I/O error:", err)
}
```

## Handling Cross-Implementation Differences

### For Users: Defensive Programming

When using ZeroKV without a specific implementation in mind, use defensive patterns:

```go
// Safe way to handle batch reuse
var batch zerokv.Batch

func useBatch(db zerokv.Core) {
    // Always create a new batch for each operation set
    batch := db.Batch()
    
    batch.Put([]byte("key1"), []byte("value1"))
    batch.Commit(context.Background())
    
    // Do NOT reuse the batch - create a new one
    batch = db.Batch()
    batch.Put([]byte("key2"), []byte("value2"))
    batch.Commit(context.Background())
}
```

### For Implementers: Standardization Pattern

If implementing a new backend and the underlying database panics on certain operations, catch and convert the panic to an error:

```go
package mydb

import (
    "context"
    "fmt"
)

type myBatch struct {
    batch *UnderlyingBatch
    closed bool
}

func (b *myBatch) Put(key []byte, value []byte) error {
    // Check if already committed
    if b.closed {
        return fmt.Errorf("batch already committed, create a new one")
    }
    
    // If underlying DB panics on error, catch it
    defer func() {
        if r := recover(); r != nil {
            // Don't let the panic propagate
            err := fmt.Errorf("batch error: %v", r)
            // Return error instead of panicking
        }
    }()
    
    return b.batch.Set(key, value)
}

func (b *myBatch) Commit(ctx context.Context) error {
    if b.closed {
        return fmt.Errorf("batch already committed")
    }
    
    b.closed = true
    
    defer func() {
        if r := recover(); r != nil {
            // Handle panic from commit
        }
    }()
    
    return b.batch.Write()
}
```

## Error Handling Checklist for New Implementations

When implementing a new backend, verify:

- `Core.Put()` returns errors consistently
- `Core.Get()` returns errors for missing keys and I/O errors
- `Core.Delete()` returns errors consistently
- `Core.Close()` returns errors if close fails
- `Batch.Put()` returns error (not panic) if batch is closed
- `Batch.Delete()` returns error (not panic) if batch is closed
- `Batch.Commit()` returns error (not panic) if already committed
- `Iterator.Error()` never panics (check empty slice)
- `Iterator.Release()` safely closes resources
- Context cancellation is respected in all operations
- Error messages are clear and helpful

## Context Error Handling

All operations must check context cancellation at the start:

```go
func (db *myDB) Get(ctx context.Context, key []byte) ([]byte, error) {
    // Check context first
    if err := ctx.Err(); err != nil {
        return nil, err
    }
    
    // Proceed with operation
    return db.db.Get(key), nil
}
```

## Testing Error Scenarios

### Example Test for Error Handling

```go
func TestBatchErrorHandling(t *testing.T) {
    db := setupDB(t)
    defer db.Close()
    
    batch := db.Batch()
    
    // Add some data
    err := batch.Put([]byte("key"), []byte("value"))
    if err != nil {
        t.Fatalf("Put failed: %v", err)
    }
    
    // Commit
    err = batch.Commit(t.Context())
    if err != nil {
        t.Fatalf("Commit failed: %v", err)
    }
    
    // Try to use batch after commit
    // This should either return an error or panic (implementation-specific)
    err = batch.Put([]byte("key2"), []byte("value2"))
    
    if err == nil && !isPanic() {
        t.Fatal("Expected error or panic when using committed batch")
    }
}

func isPanic() bool {
    // Implement based on your implementation
    return false
}
```

## Summary Table

| Operation | BadgerDB | PebbleDB | Notes |
| ----------- | ---------- | ---------- | ------- |
| Put after Commit | Error | Panic | Never reuse batch |
| Delete after Commit | Error | Panic | Create new batch |
| Commit after Commit | Error | Panic | Check closed state |
| Iterator.Error() panic | Never | Fixed | Safe to call |
| Get non-existent key | Error | Error | Same behavior |
| Context cancellation | Respected | Respected | Both check context |
| Close resources | Error if fails | Error if fails | Always check |

## Best Practices

1. **Never reuse batches**: Always create a new batch for each operation set

   ```go
   batch1 := db.Batch()
   batch1.Put([]byte("key1"), []byte("value1"))
   batch1.Commit(ctx)
   
   batch2 := db.Batch() // New batch!
   batch2.Put([]byte("key2"), []byte("value2"))
   batch2.Commit(ctx)
   ```

2. **Always check iterator.Error()**: Errors might occur during iteration

   ```go
   iter := db.Scan(prefix)
   defer iter.Release()
   for iter.Next() { /* ... */ }
   if iter.Error() != nil { /* handle */ }
   ```

3. **Check context errors**: Operations might be cancelled

   ```go
   value, err := db.Get(ctx, key)
   if errors.Is(err, context.Cancelled) {
       log.Println("Operation was cancelled")
   }
   ```

4. **Handle implementation differences**: If supporting multiple databases,
   treat error conditions consistently at your application layer

   ```go
   func storeData(db zerokv.Core, data map[string]string) error {
       batch := db.Batch()
       for k, v := range data {
           if err := batch.Put([]byte(k), []byte(v)); err != nil {
               return fmt.Errorf("batch put failed: %w", err)
           }
       }
       return batch.Commit(context.Background())
   }
   ```
