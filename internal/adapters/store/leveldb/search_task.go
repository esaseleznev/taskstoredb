package leveldb

import (
	"encoding/json"
	"fmt"

	common "github.com/esaseleznev/taskstoredb/internal/adapters/store/common"
	"github.com/esaseleznev/taskstoredb/internal/contract"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func (l LevelAdapter) SearchTask(
	condition *contract.Condition,
	kind *string,
	size *uint,
) (tasks []contract.Task, err error) {
	return l.searchTask(condition, common.PrefixTask, kind, size)
}

func (l LevelAdapter) searchTask(condition *contract.Condition, prefixTask string, kind *string, size *uint) (tasks []contract.Task, err error) {
	tasks = make([]contract.Task, 0)
	prefix := prefixTask + "-"
	if kind != nil {
		prefix = prefix + *kind + "-"
	}
	r := util.BytesPrefix([]byte(prefix))
	iter := l.db.NewIterator(r, nil)

out:
	for iter.Next() {
		task := contract.Task{}
		err := json.Unmarshal(iter.Value(), &task)
		if err != nil {
			return tasks, fmt.Errorf("task unmarshal error: %v", err)
		}
		if condition == nil || common.ConditionCalculateTask(&task, condition) {
			task.Id = string(iter.Key())
			tasks = append(tasks, task)
			if size != nil {
				*size--
			}
		}
		if size != nil && *size == 0 {
			break out
		}
	}

	iter.Release()
	err = iter.Error()

	return tasks, err
}
