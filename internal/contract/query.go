package contract

type Select struct {
	From  string
	Where Condition
}

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
	Operator Operator
	Value    any
	Field    string
}

type Condition struct {
	Operator   *LogicalOperator
	Operations []Operation
	Conditions []Condition
}
