package adapters

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"testing"

	bolt "go.etcd.io/bbolt"
)

func TestBoltRepository_OwnerReg(t *testing.T) {
	path, db, repository, err := initBoltDb()
	defer os.RemoveAll(path)
	if err != nil {
		t.Fatal(err)
	}

	kinds := []string{"one", "two", "three", "fore"}
	owner := "testowner"

	err = repository.OwnerReg(owner, kinds)
	if err != nil {
		t.Fatal(err)
	}

	err = db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(prefixOwner))
		for _, itr := range kinds {
			v := c.Get([]byte(itr + "-" + owner))
			if v == nil {
				t.Errorf("owner kind %s persist error", itr)
				return nil
			}
			var ret uint64
			binary.Read(bytes.NewBuffer(v), binary.BigEndian, &ret)
			fmt.Printf("The owner is: %v\n", ret)
		}
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}

}

func TestBoltRepository_Add(t *testing.T) {
	path, db, repository, err := initBoltDb()
	defer os.RemoveAll(path)
	if err != nil {
		t.Fatal(err)
	}

	groupIn := "12345"

	_ = repository.OwnerReg("100", []string{"TEST"})
	_ = repository.OwnerReg("101", []string{"TEST"})
	_ = repository.OwnerReg("102", []string{"TEST"})
	_ = repository.OwnerReg("103", []string{"TEST"})

	for i := 1; i < 5; i++ {
		id, err := repository.Add(groupIn, "TEST", map[string]string{"pid": groupIn, "status": "dead"})
		if err != nil {
			t.Fatal(err)
		}
		err = db.View(func(tx *bolt.Tx) error {
			c := tx.Bucket([]byte(prefixTask))
			v := c.Get([]byte(id))
			if v == nil {
				t.Errorf("not correct add task")
			}

			return nil
		})
		if err != nil {
			t.Errorf("not correct add task")
		}
	}

}

func TestBoltRepository_GetFirstInGroup(t *testing.T) {
	path, _, repository, err := initBoltDb()
	defer os.RemoveAll(path)
	if err != nil {
		t.Fatal(err)
	}

	groupIn := "12345"

	_ = repository.OwnerReg("100", []string{"TEST"})

	idIn, err := repository.Add(groupIn, "TEST", map[string]string{"pid": groupIn, "status": "dead"})
	if err != nil {
		t.Fatal(err)
	}

	for i := 1; i < 3; i++ {
		_, err := repository.Add(groupIn, "TEST", map[string]string{"pid": groupIn, "status": "dead"})
		if err != nil {
			t.Fatal(err)
		}
	}

	idOut, err := repository.GetFirstInGroup(groupIn)
	if err != nil {
		t.Fatal(err)
	}

	if idIn != idOut {
		t.Errorf("not correct first in group")
	}
}

func initBoltDb() (path string, db *bolt.DB, repository *BoltRepository, err error) {
	path = tempfile("bolt")
	db, err = bolt.Open(path, 0o600, nil)
	repository = NewBoltRepository(db)
	return path, db, repository, err
}
