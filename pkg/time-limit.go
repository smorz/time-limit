package timelimit

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

func (db *DB) SetBaseTime(t time.Time) error {
	return db.b.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte("base_time"), []byte(t.Format(layout)))
		return err
	})
}

func (db *DB) GetBaseTime() (t time.Time, err error) {
	err = db.b.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("base_time"))
		if err != nil {
			if strings.Contains(err.Error(), "Key not found") {
				t = time.Now()
				db.SetBaseTime(t)
				return nil
			} else {
				return err
			}
		}
		err = item.Value(func(val []byte) error {
			// This func with val would only be called if item.Value encounters no error.
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

func (db *DB) SetTotalDuration(d time.Duration) error {
	return db.b.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte("total_duration"), []byte(d.String()))
		return err
	})
}

func (db *DB) GetTotalDuration() (d time.Duration, err error) {
	err = db.b.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("total_duration"))
		if err != nil {
			if strings.Contains(err.Error(), "Key not found") {
				d = 0
				db.SetTotalDuration(d)
				return nil
			} else {
				return err
			}
		}
		err = item.Value(func(val []byte) error {
			// This func with val would only be called if item.Value encounters no error.
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

func (db *DB) IncTotalDuration(d time.Duration) error {
	d0, err := db.GetTotalDuration()
	if err != nil {
		return err
	}
	d0 += d
	return db.SetTotalDuration(d0)
}
func (db *DB) Close() error {
	return db.b.Close()
}
