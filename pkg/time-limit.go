//A simple wrapper for the badger
package database

import (
	"log"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v3"
)

const (
	layout = time.RFC3339
)

type DB struct {
	b *badger.DB
}

// NewDB Create a new badger Database Object
func NewDB(dir string) (*DB, error) {
	var db DB
	option := badger.DefaultOptions(dir)
	option.Logger = nil
	bdb, err := badger.Open(option)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	db.b = bdb
	return &db, nil
}

// SetTime Sets a key/value pair that the value type is Time.
func (db *DB) SetTime(key string, t time.Time) error {
	return db.b.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(key), []byte(t.Format(layout)))
		return err
	})
}

// GetTime Retrieves the value of a key and tries to convert it to Time.
func (db *DB) GetTime(key string) (t time.Time, err error) {
	err = db.b.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if strings.Contains(err.Error(), "Key not found") {
				t = time.Now()
				db.SetTime(key, t)
				return nil
			} else {
				return err
			}
		}
		err = item.Value(func(val []byte) error {
			t, err = time.Parse(layout, string(val))
			if err != nil {
				return err
			}
			return nil
		})
		return nil
	})
	return
}

// SetTime Sets a key/value pair that the value type is Duration.
func (db *DB) SetDuration(key string, d time.Duration) error {
	return db.b.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(key), []byte(d.String()))
		return err
	})
}

// GetDuration Retrieves the value of a key and tries to convert it to Duration.
func (db *DB) GetDuration(key string) (d time.Duration, err error) {
	err = db.b.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if strings.Contains(err.Error(), "Key not found") {
				d = 0
				db.SetDuration(key, d)
				return nil
			} else {
				return err
			}
		}
		err = item.Value(func(val []byte) error {
			d, err = time.ParseDuration(string(val))
			if err != nil {
				return err
			}
			return nil
		})
		return nil
	})
	return
}

// IncDuration Takes the value of a key, assuming it is of type Duration, and increase it
func (db *DB) IncDuration(key string, d time.Duration) error {
	d0, err := db.GetDuration(key)
	if err != nil {
		return err
	}
	d0 += d
	return db.SetDuration(key, d0)
}
func (db *DB) Close() error {
	return db.b.Close()
}
