package leveldb

import (
	"os"
	"strings"
	"testing"

	common "github.com/esaseleznev/taskstoredb/internal/adapters/store/common"
	"github.com/esaseleznev/taskstoredb/internal/contract"
	level "github.com/syndtr/goleveldb/leveldb"
)

func TestLevelAdapter_Add(t *testing.T) {
	path, db, adapter, err := initLevelDb()
	defer os.RemoveAll(path)
	if err != nil {
		t.Fatal(err)
	}

	groupIn := "12345"

	_ = adapter.OwnerReg("100", []string{"TEST"})
	_ = adapter.OwnerReg("101", []string{"TEST"})
	_ = adapter.OwnerReg("102", []string{"TEST"})
	_ = adapter.OwnerReg("103", []string{"TEST"})

	for i := 1; i < 5; i++ {
		p, err := adapter.Add(groupIn, "TEST", nil, map[string]string{"pid": groupIn, "status": "dead"})
		if err != nil {
			t.Fatal(err)
		}
		err = adapter.Apply(p)
		if err != nil {
			t.Fatal(err)
		}

		id := string(p[0].Key)

		_, err = db.Get([]byte(id), nil)
		if err != nil {
			t.Errorf("not correct add task")
		}
	}
}

func TestLevelAdapter_Pool(t *testing.T) {
	path, _, adapter, err := initLevelDb()
	defer os.RemoveAll(path)
	if err != nil {
		t.Fatal(err)
	}

	groupIn := "12345"

	_ = adapter.OwnerReg("100", []string{"TEST"})
	var id string
	for i := 1; i < 5; i++ {
		p, err := adapter.Add(groupIn, "TEST", nil, map[string]string{"pid": groupIn, "status": "dead"})
		if err != nil {
			t.Fatal(err)
		}
		err = adapter.Apply(p)
		if err != nil {
			t.Fatal(err)
		}

		id = string(p[0].Key)
	}

	for i := 1; i < 5; i++ {
		p, err := adapter.Add(groupIn, "TEST", nil, map[string]string{"pid": groupIn, "status": "dead"})
		if err != nil {
			t.Fatal(err)
		}
		adapter.Apply(p)
		if err != nil {
			t.Fatal(err)
		}
	}

	p, err := adapter.Update(id, contract.SCHEDULED, map[string]string{"pid": groupIn, "status": "dead"}, nil, &id)
	if err != nil {
		t.Fatal(err)
	}
	adapter.Apply(p)

	tasks, err := adapter.Pool("100", "TEST", 5)
	if err != nil {
		t.Fatal(err)
	}

	if len(tasks) != 5 {
		t.Errorf("not correct return count tasks %d", len(tasks))
	}

	if tasks[0].Id != id {
		t.Errorf("not correct return startId tasks")
	}

	for _, ts := range tasks {
		if ts.Kind != "TEST" {
			t.Errorf("not correct return kind tasks")
		}
		if *ts.Owner != "100" {
			t.Errorf("not correct return owner tasks")
		}
	}
}

func TestLevelAdapter_UpdateFailed(t *testing.T) {
	path, _, adapter, err := initLevelDb()
	defer os.RemoveAll(path)
	if err != nil {
		t.Fatal(err)
	}

	groupIn := "12345"

	_ = adapter.OwnerReg("100", []string{"TEST"})
	p, err := adapter.Add(groupIn, "TEST", nil, map[string]string{"pid": groupIn, "status": "dead"})
	if err != nil {
		t.Fatal(err)
	}

	err = adapter.Apply(p)
	if err != nil {
		t.Fatal(err)
	}

	id := string(p[0].Key)

	errorTxt := "error test"
	p, err = adapter.Update(id, contract.FAILED, map[string]string{"pid": groupIn, "status": "dead"}, &errorTxt, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = adapter.Apply(p)
	if err != nil {
		t.Fatal(err)
	}
	id = strings.Replace(
		id,
		common.PrefixTask,
		common.PrefixError,
		1,
	)
	task, err := adapter.Get(id)
	if err != nil || task == nil {
		t.Errorf("not correct update fail task")
	}
}

func TestLevelAdapter_GetFirstInGroup(t *testing.T) {
	path, _, adapter, err := initLevelDb()
	defer os.RemoveAll(path)
	if err != nil {
		t.Fatal(err)
	}

	groupIn := "12345"

	p, err := adapter.Add(groupIn, "TEST", nil, map[string]string{"pid": groupIn, "status": "dead"})
	if err != nil {
		t.Fatal(err)
	}

	err = adapter.Apply(p)
	if err != nil {
		t.Fatal(err)
	}

	idIn := string(p[0].Key)

	_, err = adapter.Add(groupIn, "TEST", nil, map[string]string{"pid": groupIn, "status": "dead"})
	if err != nil {
		t.Fatal(err)
	}

	idOut, err := adapter.GetFirstInGroup(groupIn)
	if err != nil {
		t.Fatal(err)
	}

	if idIn != idOut {
		t.Errorf("not correct first in group")
	}
}

func initLevelDb() (
	path string,
	db *level.DB,
	adapter *LevelAdapter,
	err error,
) {
	path, err = common.Tempfile("leveldb")
	if err != nil {
		return "", nil, nil, err
	}

	db, err = level.OpenFile(path, nil)
	if err != nil {
		return "", nil, nil, err
	}

	adapter, err = NewLevelAdapter(db)
	return path, db, adapter, err
}
