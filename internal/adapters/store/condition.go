package adapters

import (
	"cmp"
	"strconv"
	"strings"
	"time"

	"github.com/esaseleznev/taskstoredb/internal/contract"
)

func ConditionCalculateTask(
	task *contract.Task,
	condition *contract.Condition,
) bool {
	if condition == nil || task == nil {
		return false
	}
	res := true
	if len(condition.Operations) != 0 {
		for _, operation := range condition.Operations {
			calcRes := operationCalculateTask(task, &operation)
			if condition.LogicalOperator == nil || *condition.LogicalOperator == contract.And {
				res = res && calcRes
			} else if *condition.LogicalOperator == contract.Or {
				res = res || calcRes
			}
		}
	}

	if len(condition.Conditions) != 0 {
		for _, condition := range condition.Conditions {
			calcRes := ConditionCalculateTask(task, &condition)
			if condition.LogicalOperator == nil || *condition.LogicalOperator == contract.And {
				res = res && calcRes
			} else if *condition.LogicalOperator == contract.Or {
				res = res || calcRes
			}
		}
	}
	return res
}

func operationCalculateTask(
	task *contract.Task,
	operation *contract.Operation,
) bool {
	value := getValueTask(task, operation.Field)
	switch operation.Operator {
	case contract.Equal:
		return compare(value, operation.Value) == 0
	case contract.NotEqual:
		r := compare(value, operation.Value)
		return r == 1 || r == -1
	case contract.LessThan:
		return compare(value, operation.Value) < 0
	case contract.LessThanOrEqual:
		return compare(value, operation.Value) <= 0
	case contract.GreaterThan:
		return compare(value, operation.Value) == 1
	case contract.GreaterThanOrEqual:
		r := compare(value, operation.Value)
		return r == 1 || r == 0
	default:
		return false
	}
}

func getValueTask(task *contract.Task, field string) any {
	switch field {
	case "id":
		return task.Id
	case "kind":
		return task.Kind
	case "group":
		return task.Group
	case "owner":
		if task.Owner != nil {
			return *task.Owner
		} else {
			return nil
		}
	case "status":
		return int(task.Status)
	case "ts":
		return task.Ts
	case "error":
		if task.Error != nil {
			return *task.Error
		} else {
			return nil
		}

	default:
		s := strings.Split(field, ".")
		if len(s) > 1 && s[0] == "param" {
			return task.Param[s[1]]
		}
		return ""
	}
}

// compare any type int, float, string, bool, time.Time result 0 1 -1 2
// 2 - undefined behavior
// first parameter a this value from Task
// second parameter b this value from Condition
func compare(a, b any) int {
	if a == nil {
		if b == nil {
			return 0
		}
		return -1
	}
	switch a.(type) {
	case int:
		switch b.(type) {
		case string:
			bInt, _ := strconv.Atoi(b.(string))
			return cmp.Compare(a.(int), bInt)
		case int:
			return cmp.Compare(a.(int), b.(int))
		default:
			return 2
		}
	case float64:
		switch b.(type) {
		case string:
			bFloat, _ := strconv.ParseFloat(b.(string), 64)
			return cmp.Compare(a.(float64), bFloat)
		case float64:
			return cmp.Compare(a.(float64), b.(float64))
		default:
			return 2
		}
	case string:
		switch b.(type) {
		case string:
			return cmp.Compare(a.(string), b.(string))
		default:
			return 2
		}
	case bool:
		switch b.(type) {
		case string:
			bBool, _ := strconv.ParseBool(b.(string))
			return compareBool(a.(bool), bBool)
		case bool:
			return compareBool(a.(bool), b.(bool))
		default:
			return 2
		}
	case time.Time:
		switch b.(type) {
		case string:
			bTime, _ := time.Parse(time.RFC3339, b.(string))
			return compareTime(a.(time.Time), bTime)
		case time.Time:
			return compareTime(a.(time.Time), b.(time.Time))
		default:
			return 2
		}
	default:
		return 2
	}
}

func compareBool(a, b bool) int {
	if !a && b {
		return -1
	} else if a && !b {
		return 1
	}
	return 0
}

func compareTime(a, b time.Time) int {
	if a.Before(b) {
		return -1
	} else if a.After(b) {
		return 1
	}
	return 0
}
