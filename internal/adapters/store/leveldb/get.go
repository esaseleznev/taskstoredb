package leveldb

import (
	"encoding/json"
	"fmt"

	"github.com/esaseleznev/taskstoredb/internal/contract"
	"github.com/syndtr/goleveldb/leveldb/errors"
)

func (l LevelAdapter) Get(id string) (tasks *contract.Task, err error) {
	v, err := l.db.Get([]byte(id), nil)
	if err == errors.ErrNotFound {
		return
	}
	if err != nil {
		return nil, fmt.Errorf("get tast from db error: %v", err)
	}

	task := contract.Task{}
	err = json.Unmarshal(v, &task)
	if err != nil {
		return nil, fmt.Errorf("task unmarshal error: %v", err)
	}

	task.Id = id

	return &task, err
}
