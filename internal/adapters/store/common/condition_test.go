package common

import (
	"testing"
	"time"

	"github.com/esaseleznev/taskstoredb/internal/contract"
)

func TestCondition_Compare(t *testing.T) {
	var ts = time.Now()
	ts, _ = time.Parse(time.RFC3339, ts.Format(time.RFC3339))

	var tests = []struct {
		x, y    any
		compare int
	}{
		{1, 2, -1},
		{1, 1, 0},
		{2, 1, +1},
		{1, "2", -1},
		{1, "1", 0},
		{2, "1", +1},
		{"a", "aa", -1},
		{"a", "a", 0},
		{"aa", "a", +1},
		{1.0, 1.1, -1},
		{1.1, 1.1, 0},
		{1.1, 1.0, +1},
		{1.0, "1.1", -1},
		{1.1, "1.1", 0},
		{1.1, "1.0", +1},
		{ts.Add(-time.Second), ts, -1},
		{ts, ts, 0},
		{ts.Add(time.Second), ts, +1},
		{ts.Add(-time.Second), ts.Format(time.RFC3339), -1},
		{ts, ts, 0},
		{ts.Add(time.Second), ts.Format(time.RFC3339), +1},
		{1, ts, 2},
		{1.1, ts, 2},
		{"a", ts, 2},
		{ts, 1, 2},
	}
	for _, test := range tests {
		r := compare(test.x, test.y)
		if r != test.compare {
			t.Errorf("Expected %d, got %d", test.compare, r)
		}
	}

}

func TestCondition_ConditionCalculate(t *testing.T) {
	owner := "user"
	task := contract.Task{
		Id:     "t-test-00Q0P8XD40001",
		Kind:   "test",
		Group:  "1000",
		Owner:  &owner,
		Status: contract.VIRGIN,
		Ts:     time.Now(),
		Param:  map[string]string{"snils": "1234567890"},
		Error:  nil,
	}

	operator := contract.And
	trueCondition := contract.Condition{
		LogicalOperator: &operator,
		Operations: []contract.Operation{
			{
				Field:    "id",
				Value:    "t-test-00Q0P8XD40001",
				Operator: contract.Equal,
			},
			{
				Field:    "kind",
				Value:    "test",
				Operator: contract.Equal,
			},
			{
				Field:    "group",
				Value:    "1000",
				Operator: contract.Equal,
			},
			{
				Field:    "owner",
				Value:    owner,
				Operator: contract.Equal,
			},
			{
				Field:    "status",
				Value:    1,
				Operator: contract.Equal,
			},
			{
				Field:    "ts",
				Value:    task.Ts,
				Operator: contract.Equal,
			},
			{
				Field:    "param.snils",
				Value:    "1234567890",
				Operator: contract.Equal,
			},
			{
				Field:    "error",
				Value:    nil,
				Operator: contract.Equal,
			},
		},
	}

	res := ConditionCalculateTask(&task, &trueCondition)
	if res != true {
		t.Errorf("Expected true, got %v", res)
	}

	// Test with an empty condition
	emptyCondition := contract.Condition{
		LogicalOperator: &operator,
		Operations:      []contract.Operation{},
	}

	res = ConditionCalculateTask(&task, &emptyCondition)
	if res != true {
		t.Errorf("Expected true, got %v", res)
	}

	// Test with a condition that always evaluates to false
	falseCondition := contract.Condition{
		LogicalOperator: &operator,
		Operations: []contract.Operation{
			{
				Field:    "id",
				Value:    "t-test-00Q0P8XD40001",
				Operator: contract.NotEqual,
			},
		},
	}

	res = ConditionCalculateTask(&task, &falseCondition)
	if res != false {
		t.Errorf("Expected false, got %v", res)
	}
}
