package boltdb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	common "github.com/esaseleznev/taskstoredb/internal/adapters/store/common"
	"github.com/esaseleznev/taskstoredb/internal/contract"
	bolt "go.etcd.io/bbolt"
)

type BoltAdapter struct {
	db    *bolt.DB
	kinds map[string]*common.RoundRobin
}

func NewBoltAdapter(db *bolt.DB) *BoltAdapter {
	if db == nil {
		panic("missing db")
	}

	return &BoltAdapter{db: db, kinds: make(map[string]*common.RoundRobin)}
}

func (b BoltAdapter) GetFirstInGroup(group string) (id string, err error) {
	err = b.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(common.PrefixGroup)).Cursor()
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
	err = b.db.Update(func(tx *bolt.Tx) error {
		bktOwner, err := tx.CreateBucketIfNotExists([]byte(common.PrefixOwner))
		if err != nil {
			return fmt.Errorf("create bucket 'owner' error: %v", err)
		}

		for _, itr := range kinds {
			err = bktOwner.Put([]byte(itr+"-"+owner), nil)
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
		rr = common.NewRoundRobind(owners...)
		b.kinds[kind] = rr
	}

	owner := rr.Get()
	task := contract.Task{
		Kind:   kind,
		Group:  group,
		Param:  param,
		Status: contract.VIRGIN,
		Owner:  owner,
		Ts:     time.Now(),
	}

	taskBytes, err := json.Marshal(task)
	if err != nil {
		return id, fmt.Errorf("taskNew marshal error: %v", err)
	}

	err = b.db.Update(func(tx *bolt.Tx) error {
		bktTsk, err := tx.CreateBucketIfNotExists([]byte(common.PrefixTask))
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

		bktGrp, err := tx.CreateBucketIfNotExists([]byte(common.PrefixGroup))
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
		c := tx.Bucket([]byte(common.PrefixOwner)).Cursor()
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
