package command

import (
	"encoding/json"
	"time"

	"github.com/esaseleznev/taskstoredb/internal/contract"
	"github.com/hashicorp/raft"
)

const (
	raftTimeout = 10 * time.Second
)

type dbApply interface {
	Apply(events []contract.Event) error
}

func raftApply(raft *raft.Raft, db dbApply, events []contract.Event) error {
	if raft != nil {
		b, err := json.Marshal(events)
		if err != nil {
			return err
		}

		f := raft.Apply(b, raftTimeout)
		return f.Error()
	} else {
		return db.Apply(events)
	}
}
