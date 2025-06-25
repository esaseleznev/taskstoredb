package leveldb

import (
	"fmt"
	"time"

	common "github.com/esaseleznev/taskstoredb/internal/adapters/store/common"
	"github.com/esaseleznev/taskstoredb/internal/contract"
	"github.com/syndtr/goleveldb/leveldb"
)

func (l LevelAdapter) HealthCheck(db *leveldb.DB) (events []contract.Event, err error) {
	if l.db == nil {
		return nil, fmt.Errorf("leveldb is not initialized")
	}

	payload := common.NewPlayload()
	payload.Put([]byte("healthcheck"), []byte(time.Now().String()))

	return payload.Data(), nil
}
