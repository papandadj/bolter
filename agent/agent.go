package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/papandadj/bolter/agent/bolt"
)

var (
	errInputFormat            = errors.New("input format error")
	errInputCommandNotSupport = errors.New("command not support")
	currentBucket             = ""
)

func intro() {
	println()
	println("<------------------------------------------------------------------------->")
	println("<-----------  input `stop` or `quit` terminal process -------------------->")
	println("<-----------  buckets                           (buckets) list all bucket->")
	println("<-----------  use                          (use testbucket) select bucket->")
	println("<-----------  keys                    (keys) list all if testkey is space->")
	println("<-----------  get                                      (get key) get key->")
	println("<------------------------------------------------------------------------->")
}

func main() {
	var filePath string
	println("connecting boltdb file")
	if len(os.Args) == 2 {
		filePath = os.Args[1]
	} else {
		println("err: not  found boltdb file")
		return
	}

	db, err := bolt.Open(filePath, 0600, &bolt.Options{ReadOnly: true, Timeout: 2 * time.Second})
	if err != nil {
		println(err)
		return
	}
	defer db.Close()
	println("connectd boltdb file")

	var quit bool
	intro()
	for !quit {
		print(">>> ")
		var input string
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			input = scanner.Text()
		}

		if input == "stop" || input == "quit" {
			quit = true
			continue
		}

		err := processInput(db, input)
		if err != nil {
			println(err)
		}
	}
}

func processInput(db *bolt.DB, input string) error {
	inps := fields(input)
	inpsLen := len(inps)
	if inpsLen == 0 {
		return errInputFormat
	}

	if inps[0] == "buckets" {
		buckets(db)
	} else if inps[0] == "use" {
		if inpsLen != 2 {
			return errInputFormat
		}
		userbucket(db, inps[1])
	} else if inps[0] == "keys" {
		keys(db)
	} else if inps[0] == "get" {
		if inpsLen != 2 {
			return errInputFormat
		}
		get(db, inps[1])
	} else {
		return errInputCommandNotSupport
	}
	return nil
}

func userbucket(db *bolt.DB, name string) error {
	if !checkBucketExist(db, name) {
		return bolt.ErrBucketNotFound
	}
	currentBucket = name
	return nil
}

func checkBucketExist(db *bolt.DB, name string) bool {
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(name))
		if b == nil {
			return bolt.ErrBucketNotFound
		}
		return nil
	})

	if err != nil {
		printTxError(err)
		return false
	}

	return true
}

func keys(db *bolt.DB) {
	keys := make([]string, 0)
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(currentBucket))
		if b == nil {
			return bolt.ErrBucketNotFound
		}
		err := b.ForEach(func(k, v []byte) error {
			keys = append(keys, string(k))
			return nil
		})
		return err
	})

	if err != nil {
		printTxError(err)
		return
	}

	printArray(keys)
}

func buckets(db *bolt.DB) {
	buckets := make([]string, 0)
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Cursor().Bucket()
		err := bucket.ForEach(func(k, v []byte) error {
			buckets = append(buckets, string(k))
			return nil
		})
		return err
	})
	if err != nil {
		printTxError(err)
		return
	}
	printArray(buckets)
}

func get(db *bolt.DB, key string) {
	var val string
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(currentBucket))
		if b == nil {
			return bolt.ErrBucketNotFound
		}
		v := b.Get([]byte(key))
		val = string(v)
		return nil
	})

	if err != nil {
		printTxError(err)
		return
	}
	println(val)
}

func fields(input string) []string {
	return strings.Fields(input)
}
func printTxError(a ...interface{}) {
	fmt.Print("tx error: ")
	fmt.Println(a...)
}

func println(a ...interface{}) {
	fmt.Println(a...)
}

func print(a ...interface{}) {
	fmt.Print(a...)
}

func printArray(arr []string) {
	for index, item := range arr {
		fmt.Printf(`%d) %s`, index, item)
		fmt.Println()
	}
}
