package leveldb

import (
	"errors"

	common "github.com/esaseleznev/taskstoredb/internal/adapters/store/common"
	level "github.com/syndtr/goleveldb/leveldb"
)

type LevelAdapter struct {
	db    *level.DB
	tsid  *common.Tsid
	kinds map[string]*common.RoundRobin
}

func NewLevelAdapter(db *level.DB) (*LevelAdapter, error) {
	if db == nil {
		return nil, errors.New("missing db")
	}

	return &LevelAdapter{db: db, kinds: make(map[string]*common.RoundRobin), tsid: common.NewTsid()}, nil
}
