package leveldb

import (
	"fmt"

	common "github.com/esaseleznev/taskstoredb/internal/adapters/store/common"
	level "github.com/syndtr/goleveldb/leveldb"
)

func (l LevelAdapter) OwnerReg(owner string, kinds []string) (err error) {
	batch := new(level.Batch)
	for _, itr := range kinds {
		keyOwner := fmt.Sprintf("%s-%s-%s", common.PrefixOwner, itr, owner)
		batch.Put([]byte(keyOwner), nil)
	}
	err = l.db.Write(batch, nil)
	if err != nil {
		return fmt.Errorf("could not set owner in bucket 'owner': %v", err)
	}

	return err
}
