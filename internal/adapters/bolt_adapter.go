package adapters

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	entity "github.com/esaseleznev/taskstoredb/internal/domain"
	bolt "go.etcd.io/bbolt"
)

type BoltAdapter struct {
	db    *bolt.DB
	kinds map[string]*roundRobin
}

func NewBoltAdapter(db *bolt.DB) *BoltAdapter {
	if db == nil {
		panic("missing db")
	}

	return &BoltAdapter{db: db, kinds: make(map[string]*roundRobin)}
}

func (b BoltAdapter) GetFirstInGroup(group string) (id string, err error) {
	err = b.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(prefixGroup)).Cursor()
		prefix := []byte(group + "-")
		k, _ := c.Seek(prefix)
		if k != nil {
			slice := strings.Split(string(k), "-")
			if len(slice) == 2 {
				id = slice[1]
			}
		}
		return nil
	})

	return id, err
}

func (b BoltAdapter) OwnerReg(owner string, kinds []string) (err error) {
	offset := Offset{
		Value: 0,
		Ts:    time.Now(),
	}

	offsetBytes, err := json.Marshal(offset)
	if err != nil {
		return fmt.Errorf("offset marshal error: %v", err)
	}

	err = b.db.Update(func(tx *bolt.Tx) error {
		bktOwner, err := tx.CreateBucketIfNotExists([]byte(prefixOwner))
		if err != nil {
			return fmt.Errorf("create bucket 'owner' error: %v", err)
		}

		for _, itr := range kinds {
			err = bktOwner.Put([]byte(itr+"-"+owner), offsetBytes)
			if err != nil {
				return fmt.Errorf("could not set owner in bucket 'owner': %v", err)
			}
		}

		return nil
	})

	return err
}

func (b BoltAdapter) Add(group string, kind string, param map[string]string) (id string, err error) {
	rr, ok := b.kinds[kind]
	if !ok {
		owners, err := b.getOwnersKind(kind)
		if err != nil {
			return id, err
		}
		rr = newRoundRobind(owners...)
		b.kinds[kind] = rr
	}

	owner := rr.get()
	task := entity.Task{
		Kind:   kind,
		Group:  group,
		Param:  param,
		Status: entity.VIRGIN,
		Owner:  owner,
		Ts:     time.Now(),
	}

	taskBytes, err := json.Marshal(task)
	if err != nil {
		return id, fmt.Errorf("taskNew marshal error: %v", err)
	}

	err = b.db.Update(func(tx *bolt.Tx) error {
		bktTsk, err := tx.CreateBucketIfNotExists([]byte(prefixTask))
		if err != nil {
			return fmt.Errorf("create bucket 'task' error: %v", err)
		}

		num, err := bktTsk.NextSequence()
		if err != nil {
			return fmt.Errorf("next sequence bucket 'task' error: %v", err)
		}

		id = strconv.Itoa(int(num))

		err = bktTsk.Put([]byte(id), taskBytes)
		if err != nil {
			return fmt.Errorf("could not set task in bucket 'task': %v", err)
		}

		bktGrp, err := tx.CreateBucketIfNotExists([]byte(prefixGroup))
		if err != nil {
			return fmt.Errorf("next sequence bucket 'group' error: %v", err)
		}

		err = bktGrp.Put([]byte(group+"-"+id), nil)
		if err != nil {
			return fmt.Errorf("could not set group in bucket 'group': %v", err)
		}

		return nil
	})

	return id, err
}

func (b BoltAdapter) getOwnersKind(kind string) (owners []string, err error) {
	prefix := []byte(kind + "-")
	owners = []string{}
	err = b.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(prefixOwner)).Cursor()
		for k, _ := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, _ = c.Next() {
			its := strings.Split(string(k), "-")
			if len(its) == 2 {
				owners = append(owners, its[1])
			}
		}
		return nil
	})

	return owners, err
}
