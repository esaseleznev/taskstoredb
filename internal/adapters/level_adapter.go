package adapters

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/esaseleznev/taskstoredb/internal/contract"
	leveldb "github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type LevelAdapter struct {
	db    *leveldb.DB
	tsid  *tsid
	kinds map[string]*roundRobin
}

func NewLevelAdapter(db *leveldb.DB) *LevelAdapter {
	if db == nil {
		panic("missing db")
	}

	return &LevelAdapter{db: db, kinds: make(map[string]*roundRobin), tsid: newTsid()}
}

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

func (l LevelAdapter) Pool(owner string, kind string, size uint) (tasks []contract.Task, err error) {
	tasks = make([]contract.Task, 0)
	prefix := prefixTask + "-" + kind + "-"
	r := util.BytesPrefix([]byte(prefix))

	keyOffset := fmt.Sprintf("%s-%s-%s", prefixOffset, owner, kind)
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

func (l LevelAdapter) SearchTask(condition *contract.Condition, kind *string, size *uint) (tasks []contract.Task, err error) {
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
		if condition == nil || ConditionCalculateTask(&task, condition) {
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

func (l LevelAdapter) GetFirstInGroup(group string) (id string, err error) {
	prefix := prefixGroup + "-" + group + "-"
	iter := l.db.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	if iter.Next() {
		id = string(iter.Value())
	}
	iter.Release()
	err = iter.Error()

	return id, err
}

func (l LevelAdapter) OwnerReg(owner string, kinds []string) (err error) {
	batch := new(leveldb.Batch)
	for _, itr := range kinds {
		keyOwner := fmt.Sprintf("%s-%s-%s", prefixOwner, itr, owner)
		batch.Put([]byte(keyOwner), nil)
	}
	err = l.db.Write(batch, nil)
	if err != nil {
		return fmt.Errorf("could not set owner in bucket 'owner': %v", err)
	}

	return err
}

func (l LevelAdapter) OwnerUnreg(owner string) (err error) {
	prefix := prefixOwner + "-"
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

	batch := new(leveldb.Batch)
	for _, key := range keys {
		batch.Delete([]byte(key))
	}
	err = l.db.Write(batch, nil)
	if err != nil {
		return fmt.Errorf("could not delete owner keys: %v", err)
	}

	return err
}

func (l LevelAdapter) SetOffset(owner string, kind string, startId string) (err error) {
	keyOffset := fmt.Sprintf("%s-%s-%s", prefixOffset, owner, kind)
	err = l.db.Put([]byte(keyOffset), []byte(startId), nil)
	if err != nil {
		return fmt.Errorf("could not set offset: %v", err)
	}
	return
}

func (l *LevelAdapter) Add(group string, kind string, param map[string]string) (id string, err error) {
	rr, ok := l.kinds[kind]
	if !ok {
		owners, err := l.getOwnersKind(kind)
		if err != nil {
			return id, err
		}
		rr = newRoundRobind(owners...)
		l.kinds[kind] = rr
	}

	owner := rr.get()
	task := contract.Task{
		Kind:   kind,
		Group:  group,
		Param:  param,
		Status: contract.VIRGIN,
		Owner:  owner,
		Ts:     time.Now(),
	}

	taskBytes, err := json.Marshal(task)
	if err != nil {
		return id, fmt.Errorf("taskNew marshal error: %v", err)
	}

	ts := l.tsid.next(task.Ts.UnixMilli())

	id = fmt.Sprintf("%s-%s-%s", prefixTask, kind, ts)
	keyGroup := fmt.Sprintf("%s-%s-%s", prefixGroup, group, ts)
	batch := new(leveldb.Batch)
	batch.Put([]byte(id), []byte(taskBytes))
	batch.Put([]byte(keyGroup), []byte(id))
	err = l.db.Write(batch, nil)
	if err != nil {
		return id, fmt.Errorf("could not set task in bucket 'task': %v", err)
	}

	return id, err
}

func (l LevelAdapter) Update(
	id string,
	status contract.Status,
	param map[string]string,
	error *string,
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

	batch := new(leveldb.Batch)

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
				prefixTask,
				prefixError,
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
	its[0] = prefixGroup
	its[1] = group
	groupId = fmt.Sprintf("%s-%s-%s", its[0], its[1], its[2])

	return groupId, err
}

func (l LevelAdapter) getOwnersKind(kind string) (owners []string, err error) {
	prefix := prefixOwner + "-" + kind + "-"
	owners = []string{}
	iter := l.db.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	for iter.Next() {
		its := strings.Split(string(iter.Key()), "-")
		if len(its) == 3 {
			owners = append(owners, its[2])
		}
	}
	iter.Release()
	err = iter.Error()
	return owners, err
}
