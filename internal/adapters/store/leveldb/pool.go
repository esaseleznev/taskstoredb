package leveldb

import (
	"encoding/json"
	"fmt"

	common "github.com/esaseleznev/taskstoredb/internal/adapters/store/common"
	"github.com/esaseleznev/taskstoredb/internal/contract"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func (l LevelAdapter) Pool(
	owner string,
	kind string,
	size uint,
) (tasks []contract.Task, err error) {
	tasks = make([]contract.Task, 0)
	prefix := common.PrefixTask + "-" + kind + "-"
	r := util.BytesPrefix([]byte(prefix))

	keyOffset := fmt.Sprintf("%s-%s-%s", common.PrefixOffset, owner, kind)
	startId, err := l.db.Get([]byte(keyOffset), nil)
	if err != nil && err != errors.ErrNotFound {
		return tasks, fmt.Errorf("task get offset error: %v", err)
	}
	if err != errors.ErrNotFound {
		r.Start = startId
	}

	iter := l.db.NewIterator(r, nil)

out:
	for iter.Next() {
		task := contract.Task{}
		err := json.Unmarshal(iter.Value(), &task)
		if err != nil {
			return tasks, fmt.Errorf("task unmarshal error: %v", err)
		}
		if owner == *task.Owner {
			task.Id = string(iter.Key())
			tasks = append(tasks, task)
			size--
		}
		if size == 0 {
			break out
		}
	}

	iter.Release()
	err = iter.Error()

	return tasks, err
}
