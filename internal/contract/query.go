package contract

type Operator string

const (
	Equal              Operator = "="
	NotEqual           Operator = "!="
	LessThan           Operator = "<"
	LessThanOrEqual    Operator = "<="
	GreaterThan        Operator = ">"
	GreaterThanOrEqual Operator = ">="
	Contains           Operator = "LIKE"
)

type LogicalOperator string

const (
	And LogicalOperator = "AND"
	Or  LogicalOperator = "OR"
)

type Operation struct {
	Operator Operator `json:"op"`
	Value    any      `json:"val"`
	Field    string   `json:"fld"`
}

type Condition struct {
	LogicalOperator *LogicalOperator `json:"lop"`
	Operations      []Operation      `json:"ops"`
	Conditions      []Condition      `json:"conds"`
}
