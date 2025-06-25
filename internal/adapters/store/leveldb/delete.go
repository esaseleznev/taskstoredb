package leveldb

import (
	"github.com/esaseleznev/taskstoredb/internal/contract"
)

func (l LevelAdapter) Delete(
	id string,
) (events []contract.Event, err error) {
	return l.Update(id, contract.COMPLETED, nil, nil, nil)
}
