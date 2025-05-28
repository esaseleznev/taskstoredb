package leveldb

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	common "github.com/esaseleznev/taskstoredb/internal/adapters/store/common"
	"github.com/esaseleznev/taskstoredb/internal/contract"
	level "github.com/syndtr/goleveldb/leveldb"
)

func (l LevelAdapter) Update(
	id string,
	status contract.Status,
	param map[string]string,
	error *string,
	offset *string,

) (err error) {
	task, err := l.Get(id)
	if err != nil {
		return fmt.Errorf("get tast from db error: %v", err)
	}
	if task == nil {
		return
	}

	task.Param = param
	task.Status = status

	batch := new(level.Batch)

	switch status {
	case contract.SCHEDULED:
	case contract.VIRGIN:
		taskBytes, err := json.Marshal(task)
		if err != nil {
			return fmt.Errorf("task marshal error: %v", err)
		}
		batch.Put([]byte(id), taskBytes)
	case contract.FAILED:
		taskError := contract.Task{
			Id: strings.Replace(
				id,
				common.PrefixTask,
				common.PrefixError,
				1,
			),
			Kind:  task.Kind,
			Group: task.Group,
			Param: task.Param,
			Error: error,
			Ts:    time.Now(),
		}

		groupId, err := l.getGroupId(id, taskError.Group)
		if err != nil {
			return fmt.Errorf("taskError group parse error: %v", err)
		}
		taskBytes, err := json.Marshal(taskError)
		if err != nil {
			return fmt.Errorf("taskError marshal error: %v", err)
		}
		batch.Delete([]byte(groupId))
		batch.Delete([]byte(id))
		batch.Put([]byte(taskError.Id), taskBytes)
	case contract.COMPLETED:
		groupId, err := l.getGroupId(id, task.Group)
		if err != nil {
			return fmt.Errorf("taskError group parse error: %v", err)
		}
		batch.Delete([]byte(groupId))
		batch.Delete([]byte(id))
	default:
		return fmt.Errorf("unexpected status: %v", status)
	}

	if offset != nil && task.Owner != nil {
		keyOffset := fmt.Sprintf("%s-%s-%s", common.PrefixOffset, *task.Owner, task.Kind)
		batch.Put([]byte(keyOffset), []byte(*offset))
	}

	err = l.db.Write(batch, nil)
	if err != nil {
		return fmt.Errorf("could not update task in bucket 'task': %v", err)
	}

	return err
}

func (l LevelAdapter) getGroupId(id string, group string) (groupId string, err error) {
	its := strings.Split(string(id), "-")
	if len(its) != 3 {
		return groupId, fmt.Errorf("could not parse groupId, format error: %v", id)
	}
	its[0] = common.PrefixGroup
	its[1] = group
	groupId = fmt.Sprintf("%s-%s-%s", its[0], its[1], its[2])

	return groupId, err
}
