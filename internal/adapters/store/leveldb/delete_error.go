package leveldb

import (
	"github.com/esaseleznev/taskstoredb/internal/contract"
)

func (l LevelAdapter) DeleteError(
	id string,
) (events []contract.Event, err error) {
	return l.UpdateError(id, contract.COMPLETED, nil)

}
