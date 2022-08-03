//A simple wrapper for the badger
package database

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"time"
)

/*
const (
	layout = time.RFC3339
)
*/
type DB struct {
	f *os.File
	m map[string]interface{}
}

// OpenDB Create a new badger Database Object
func OpenDB(file string) (*DB, error) {
	var db DB
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &db.m)
	if err != nil {
		b = []byte("{}")
		err = json.Unmarshal(b, &db.m)
		if err != nil {
			return nil, err
		}
	}
	db.f = f
	return &db, nil
}

// Set Sets a key/value pair.
func (db *DB) Set(key string, t interface{}) error {
	db.m[key] = t
	b, err := json.MarshalIndent(db.m, "", "	")
	if err != nil {
		return err
	}
	err = db.f.Truncate(0)
	if err != nil {
		return err
	}
	_, err = db.f.WriteAt(b, 0)
	if err != nil {
		return err
	}
	err = db.f.Sync()
	if err != nil {
		return err
	}
	return nil
}

// GetTime Retrieves the value of a key and tries to convert it to Time.
func (db *DB) GetTime(key string) (t time.Time, err error) {
	if s, ok := db.m[key].(string); ok {
		t, err = time.Parse(time.RFC3339Nano, s)
		return
	}
	t = time.Now()
	err = db.Set(key, t)
	return
}

// GetDuration Retrieves the value of a key and tries to convert it to Duration.
func (db *DB) GetDuration(key string) (d time.Duration, err error) {
	dur, ok := db.m[key]
	if !ok {
		return 0, nil
	}
	switch dur.(type) {
	case time.Duration:
		return dur.(time.Duration), nil
	case float64:
		return time.Duration(dur.(float64)), nil
	case int:
		return time.Duration(dur.(int)), nil
	default:
		return 0, fmt.Errorf("unknown type: %v", reflect.TypeOf(dur))
	}
}

// IncDuration Takes the value of a key, assuming it is of type Duration, and increase it
func (db *DB) IncDuration(key string, d time.Duration) error {
	d0, err := db.GetDuration(key)
	if err != nil {
		return err
	}
	d0 += d
	return db.Set(key, d0)
}

// Close Closes the file.
func (db *DB) Close() error {
	err := db.f.Sync()
	if err != nil {
		return err
	}
	return db.f.Close()
}
