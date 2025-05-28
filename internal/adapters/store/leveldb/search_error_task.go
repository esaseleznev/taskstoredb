package leveldb

import (
	common "github.com/esaseleznev/taskstoredb/internal/adapters/store/common"
	"github.com/esaseleznev/taskstoredb/internal/contract"
)

func (l LevelAdapter) SearchErrorTask(condition *contract.Condition, kind *string, size *uint) (tasks []contract.Task, err error) {
	return l.searchTask(condition, common.PrefixError, kind, size)
}
