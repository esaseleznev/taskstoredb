package leveldb

import (
	"encoding/json"
	"fmt"

	"github.com/esaseleznev/taskstoredb/internal/contract"
	level "github.com/syndtr/goleveldb/leveldb"
)

func (l LevelAdapter) UpdateError(
	id string,
	status contract.Status,
	param map[string]string,

) (err error) {
	taskError, err := l.Get(id)
	if err != nil {
		return fmt.Errorf("get tast from db error: %v", err)
	}
	if taskError == nil {
		return
	}

	taskError.Param = param
	taskError.Status = status

	batch := new(level.Batch)

	switch status {
	case contract.FAILED:
		taskBytes, err := json.Marshal(taskError)
		if err != nil {
			return fmt.Errorf("task marshal error: %v", err)
		}
		batch.Put([]byte(id), taskBytes)
	case contract.VIRGIN:
	case contract.SCHEDULED:
		task, idNew, keyGroup, err := l.newTask(taskError.Group, taskError.Kind, taskError.Owner, taskError.Param)
		if err != nil {
			return err
		}

		taskBytes, err := json.Marshal(task)
		if err != nil {
			return fmt.Errorf("taskNew marshal error: %v", err)
		}

		batch.Put([]byte(idNew), []byte(taskBytes))
		batch.Put([]byte(keyGroup), []byte(idNew))
		batch.Delete([]byte(id))
	case contract.COMPLETED:
		batch.Delete([]byte(id))
	default:
		return fmt.Errorf("unexpected status: %v", status)
	}

	err = l.db.Write(batch, nil)
	if err != nil {
		return fmt.Errorf("could not update task in bucket 'task': %v", err)
	}

	return err
}
