package main

import (
	"fmt"
	"log"
	"time"

	"github.com/boltdb/bolt"
)

func main() {
	db, err := bolt.Open("example/eg.db", 0600, &bolt.Options{Timeout: 2 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	fmt.Println("connect")

	{
		//crate bucket
		tx, err := db.Begin(true)
		if err != nil {
			log.Fatal(err)
		}
		_, err = tx.CreateBucketIfNotExists([]byte("drama"))
		passRollBack(tx, err)
		_, err = tx.CreateBucketIfNotExists([]byte("music"))
		passRollBack(tx, err)
		// Commit the transaction and check for error.
		if err := tx.Commit(); err != nil {
			passRollBack(tx, err)
		}
	}

	{
		//create key val
		tx, err := db.Begin(true)
		passRollBack(tx, err)
		b := tx.Bucket([]byte("drama"))
		err = b.Put([]byte("game_of_thrones"), []byte("Game of Thrones is an American fantasy drama television series"))
		passRollBack(tx, err)
		err = b.Put([]byte("friends"), []byte("Friends tells the story of siblings Ross and Monica Geller, Chandler Bing, Phoebe Buffay, Joey Tribiani and Rachel Green"))
		passRollBack(tx, err)
		b = tx.Bucket([]byte("music"))
		err = b.Put([]byte("sia"), []byte("unstoppable"))
		passRollBack(tx, err)
		err = b.Put([]byte("katy_perry"), []byte("firework"))
		passRollBack(tx, err)
		// Commit the transaction and check for error.
		if err := tx.Commit(); err != nil {
			passRollBack(tx, err)
		}
	}
	fmt.Println("finish")
}

func passRollBack(tx *bolt.Tx, err error) {
	if err != nil {
		tx.Rollback()
		panic(err)
	}
}
