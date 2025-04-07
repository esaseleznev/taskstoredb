package adapters

import (
	"cmp"
	"strconv"
	"time"

	"github.com/esaseleznev/taskstoredb/internal/contract"
)

func ConditionCalculate(
	task *contract.Task,
	expression *contract.Condition,
) bool {
	if expression == nil || task == nil {
		return false
	}
	res := true
	if len(expression.Operations) != 0 {
		for _, operation := range expression.Operations {
			calcRes := operationCalculate(task, &operation)
			if expression.Operator == nil || *expression.Operator == contract.And {
				res = res && calcRes
			} else if *expression.Operator == contract.Or {
				res = res || calcRes
			}
		}
	}

	if len(expression.Conditions) != 0 {
		for _, condition := range expression.Conditions {
			calcRes := ConditionCalculate(task, &condition)
			if expression.Operator == nil || *expression.Operator == contract.And {
				res = res && calcRes
			} else if *expression.Operator == contract.Or {
				res = res || calcRes
			}
		}
	}
	return res
}

func operationCalculate(
	task *contract.Task,
	operation *contract.Operation,
) bool {
	value := getValue(task, operation.Field)
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

func getValue(task *contract.Task, field string) any {
	switch field {
	case "id":
		return task.Id
	case "kind":
		return task.Kind
	case "group":
		return task.Group
	case "owner":
		return *task.Owner
	case "status":
		return int(task.Status)
	case "ts":
		return task.Ts
	case "error":
		return *task.Error
	default:
		return ""
	}
}

// compare any type int, float, string, bool, time.Time result 0 1 -1 2
// 2 - undefined behavior
// first parameter a this value from Task
// second parameter b this value from Condition
func compare(a, b any) int {
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
