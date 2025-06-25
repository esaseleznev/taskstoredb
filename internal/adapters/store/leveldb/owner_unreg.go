package leveldb

import (
	"fmt"

	common "github.com/esaseleznev/taskstoredb/internal/adapters/store/common"
	"github.com/esaseleznev/taskstoredb/internal/contract"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func (l LevelAdapter) OwnerUnReg(owner string) (events []contract.Event, err error) {
	prefix := common.PrefixOwner + "-"
	var keys = []string{}
	iter := l.db.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	for iter.Next() {
		keys = append(keys, string(iter.Key()))
	}
	iter.Release()
	err = iter.Error()
	if err != nil {
		return nil, fmt.Errorf("could not get owner keys: %v", err)
	}

	payload := common.NewPlayload()
	for _, key := range keys {
		payload.Delete([]byte(key), nil)
	}

	return payload.Data(), err
}
