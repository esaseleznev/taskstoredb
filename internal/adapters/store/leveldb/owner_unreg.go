package leveldb

import (
	"fmt"

	common "github.com/esaseleznev/taskstoredb/internal/adapters/store/common"
	level "github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func (l LevelAdapter) OwnerUnreg(owner string) (err error) {
	prefix := common.PrefixOwner + "-"
	var keys = []string{}
	iter := l.db.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	for iter.Next() {
		keys = append(keys, string(iter.Key()))
	}
	iter.Release()
	err = iter.Error()
	if err != nil {
		return fmt.Errorf("could not get owner keys: %v", err)
	}

	batch := new(level.Batch)
	for _, key := range keys {
		batch.Delete([]byte(key))
	}
	err = l.db.Write(batch, nil)
	if err != nil {
		return fmt.Errorf("could not delete owner keys: %v", err)
	}

	return err
}
