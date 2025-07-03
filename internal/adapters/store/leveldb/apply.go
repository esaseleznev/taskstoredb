package leveldb

import (
	"github.com/esaseleznev/taskstoredb/internal/contract"
	level "github.com/syndtr/goleveldb/leveldb"
)

// for compatibility with Raft consensus algorithm
func (l LevelAdapter) Apply(events []contract.Event) (err error) {
	return ApplyDb(l.db, events)
}

func ApplyDb(db *level.DB, events []contract.Event) error {
	batch := new(level.Batch)
	for _, e := range events {
		switch e.Type {
		case contract.SetType:
			batch.Put(e.Key, e.Value)
		case contract.DeleteType:
			batch.Delete(e.Key)
		}
	}
	return db.Write(batch, nil)
}
