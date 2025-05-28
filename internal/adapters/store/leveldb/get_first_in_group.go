package leveldb

import (
	common "github.com/esaseleznev/taskstoredb/internal/adapters/store/common"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func (l LevelAdapter) GetFirstInGroup(group string) (id string, err error) {
	prefix := common.PrefixGroup + "-" + group + "-"
	iter := l.db.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	if iter.Next() {
		id = string(iter.Value())
	}
	iter.Release()
	err = iter.Error()

	return id, err
}
