package leveldb

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	common "github.com/esaseleznev/taskstoredb/internal/adapters/store/common"
	"github.com/esaseleznev/taskstoredb/internal/contract"
	level "github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func (l *LevelAdapter) Add(group string, kind string, owner *string, param map[string]string) (id string, err error) {
	task, id, keyGroup, err := l.newTask(group, kind, owner, param)
	if err != nil {
		return id, err
	}

	taskBytes, err := json.Marshal(task)
	if err != nil {
		return id, fmt.Errorf("taskNew marshal error: %v", err)
	}

	batch := new(level.Batch)
	batch.Put([]byte(id), []byte(taskBytes))
	batch.Put([]byte(keyGroup), []byte(id))
	err = l.db.Write(batch, nil)
	if err != nil {
		return id, fmt.Errorf("could not set task in bucket 'task': %v", err)
	}

	return id, err
}

func (l *LevelAdapter) newTask(group string, kind string, owner *string, param map[string]string) (task contract.Task, id string, keyGroup string, err error) {
	rr, ok := l.kinds[kind]
	if !ok {
		owners, err := l.getOwnersKind(kind)
		if err != nil {
			return task, id, keyGroup, err
		}
		rr = common.NewRoundRobind(owners...)
		l.kinds[kind] = rr
	}

	if owner == nil {
		owner = rr.Get()
	}
	task = contract.Task{
		Kind:   kind,
		Group:  group,
		Param:  param,
		Status: contract.VIRGIN,
		Owner:  owner,
		Ts:     time.Now(),
	}

	ts := l.tsid.Next(task.Ts.UnixMilli())

	id = fmt.Sprintf("%s-%s-%s", common.PrefixTask, kind, ts)
	keyGroup = fmt.Sprintf("%s-%s-%s", common.PrefixGroup, group, ts)

	return task, id, keyGroup, nil
}

func (l LevelAdapter) getOwnersKind(kind string) (owners []string, err error) {
	prefix := common.PrefixOwner + "-" + kind + "-"
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
