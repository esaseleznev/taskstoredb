package adapters

import (
	"os"
	"strings"
	"testing"

	entity "github.com/esaseleznev/taskstoredb/internal/domain"
	leveldb "github.com/syndtr/goleveldb/leveldb"
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
		id, err := adapter.Add(groupIn, "TEST", map[string]string{"pid": groupIn, "status": "dead"})
		if err != nil {
			t.Fatal(err)
		}
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
		id, err = adapter.Add(groupIn, "TEST", map[string]string{"pid": groupIn, "status": "dead"})
		if err != nil {
			t.Fatal(err)
		}
	}

	for i := 1; i < 5; i++ {
		_, err = adapter.Add(groupIn, "TEST1", map[string]string{"pid": groupIn, "status": "dead"})
		if err != nil {
			t.Fatal(err)
		}
	}

	for i := 1; i < 7; i++ {
		_, err = adapter.Add(groupIn, "TEST", map[string]string{"pid": groupIn, "status": "dead"})
		if err != nil {
			t.Fatal(err)
		}
	}

	err = adapter.SetOffset("TEST", id)
	if err != nil {
		t.Fatal(err)
	}

	tasks, err := adapter.Pool("100", "TEST", 5)
	if err != nil {
		t.Fatal(err)
	}

	if len(tasks) != 5 {
		t.Errorf("not correct return count tasks")
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
	id, err := adapter.Add(groupIn, "TEST", map[string]string{"pid": groupIn, "status": "dead"})
	if err != nil {
		t.Fatal(err)
	}
	errorTxt := "error test"
	err = adapter.Update(id, entity.FAILED, map[string]string{"pid": groupIn, "status": "dead"}, &errorTxt)
	if err != nil {
		t.Fatal(err)
	}
	id = strings.Replace(
		id,
		prefixTask,
		prefixError,
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

	idIn, err := adapter.Add(groupIn, "TEST", map[string]string{"pid": groupIn, "status": "dead"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = adapter.Add(groupIn, "TEST", map[string]string{"pid": groupIn, "status": "dead"})
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
	db *leveldb.DB,
	adapter *LevelAdapter,
	err error,
) {
	path = tempfile("leveldb")
	db, err = leveldb.OpenFile(path, nil)
	adapter = NewLevelAdapter(db)
	return path, db, adapter, err
}
