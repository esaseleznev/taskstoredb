package leveldb

import (
	"fmt"

	common "github.com/esaseleznev/taskstoredb/internal/adapters/store/common"
	"github.com/esaseleznev/taskstoredb/internal/contract"
)

func (l LevelAdapter) OwnerReg(owner string, kinds []string) (events []contract.Event) {
	payload := common.NewPlayload()
	for _, itr := range kinds {
		keyOwner := fmt.Sprintf("%s-%s-%s", common.PrefixOwner, itr, owner)
		payload.Put([]byte(keyOwner), nil)
	}
	events = payload.Data()

	return events
}
