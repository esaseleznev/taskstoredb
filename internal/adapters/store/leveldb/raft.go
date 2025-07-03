package leveldb

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/esaseleznev/taskstoredb/internal/contract"
	"github.com/hashicorp/raft"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/tidwall/sds"
)

type Fsm LevelAdapter

func (f *Fsm) Apply(l *raft.Log) any {
	var event contract.Event
	if err := json.Unmarshal(l.Data, &event); err != nil {
		return fmt.Errorf("event marshal error: %v", err)
	}
	if err := ApplyDb(f.db, []contract.Event{event}); err != nil {
		return fmt.Errorf("failed to apply event: %v", err)
	}
	return nil
}

func (f *Fsm) Snapshot() (raft.FSMSnapshot, error) {
	s, err := f.db.GetSnapshot()
	if err != nil {
		return nil, err
	}
	return &FsmSnapshot{store: s}, nil
}

func (f *Fsm) Restore(rc io.ReadCloser) error {
	sr := sds.NewReader(rc)
	var batch leveldb.Batch
	for {
		key, err := sr.ReadBytes()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		value, err := sr.ReadBytes()
		if err != nil {
			return err
		}
		batch.Put(key, value)
		if batch.Len() == 1000 {
			if err := f.db.Write(&batch, nil); err != nil {
				return err
			}
			batch.Reset()
		}
	}
	if err := f.db.Write(&batch, nil); err != nil {
		return err
	}
	return nil
}

type FsmSnapshot struct {
	store *leveldb.Snapshot
}

func (f *FsmSnapshot) Persist(sink raft.SnapshotSink) error {
	err := func() error {
		sw := sds.NewWriter(sink)
		iter := f.store.NewIterator(nil, nil)
		for ok := iter.First(); ok; ok = iter.Next() {
			if err := sw.WriteBytes(iter.Key()); err != nil {
				return err
			}
			if err := sw.WriteBytes(iter.Value()); err != nil {
				return err
			}
		}
		iter.Release()
		if err := iter.Error(); err != nil {
			return err
		}

		err := sw.Flush()
		if err != nil {
			return err
		}
		return sink.Close()
	}()

	if err != nil {
		sink.Cancel()
	}

	return err
}

func (f *FsmSnapshot) Release() {
	f.store.Release()
}
