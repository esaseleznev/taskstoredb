package leveldb

import (
	"github.com/esaseleznev/taskstoredb/internal/contract"
)

func (l LevelAdapter) DeleteError(
	id string,
) (err error) {
	l.UpdateError(id, contract.COMPLETED, nil)
	return err
}
