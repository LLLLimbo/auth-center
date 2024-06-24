package main

import (
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"log"
	"time"
)

type Db struct {
	Db *badger.DB
}

func InitDb() *Db {
	opt := badger.DefaultOptions("").WithInMemory(true)
	db, err := badger.Open(opt)
	if err != nil {
		log.Fatal(err)
	}
	return &Db{Db: db}
}

func (db *Db) Get(key []byte) (string, error) {
	var valCopy []byte
	err := db.Db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			log.Printf("Badger error: %s", err)
			return err
		}
		valCopy, err = item.ValueCopy(nil)
		if err != nil {
			log.Printf("Badger error: %s", err)
			return err
		}
		return nil
	})
	return string(valCopy), err
}

func (db *Db) Set(key []byte, val []byte, ttl time.Duration) error {
	err := db.Db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry(key, val).WithTTL(ttl)
		err := txn.SetEntry(e)
		if err != nil {
			log.Printf("Badger error: %s", err)
			return err
		}
		return nil
	})
	return err
}

func (db *Db) Delete(key []byte) error {
	err := db.Db.Update(func(txn *badger.Txn) error {
		err := txn.Delete(key)
		if err != nil {
			log.Printf("Badger error: %s", err)
			return err
		}
		return nil
	})
	return err
}

func (db *Db) Iterator() {
	err := db.Db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			v, _ := item.ValueCopy(nil)
			fmt.Printf("key=%s, value=%s\n", k, v)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("error while iterating: %s\n", err)
	}
}
