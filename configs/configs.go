package configs

import (
	"github.com/cockroachdb/pebble"
	"github.com/dgraph-io/badger/v4"
)

type StoreConfig struct {
	BadgerConfigs *badger.Options
	PebbleConfigs *pebble.Options
	Default       *DefaultOptions
}

type DefaultOptions struct {
	Dir string // some databases may require to specify the storage directory seperatly
}
