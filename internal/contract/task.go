package contract

import "time"

type Task struct {
	Id     string            `json:"id,omitzero"`
	Kind   string            `json:"k"`
	Group  string            `json:"g"`
	Owner  *string           `json:"o,omitzero"`
	Status Status            `json:"s"`
	Param  map[string]string `json:"p"`
	Ts     time.Time         `json:"t"`
	Error  *string           `json:"e,omitzero"`
}

type Status int

const (
	VIRGIN    Status = 1
	SCHEDULED Status = 2
	COMPLETED Status = 3
	FAILED    Status = 4
)

type TaskUpdate struct {
	Kind   *string           `json:"k"`
	Group  *string           `json:"g"`
	Owner  *string           `json:"o,omitzero"`
	Status *Status           `json:"s"`
	Param  map[string]string `json:"p"`
	Error  *string           `json:"e,omitzero"`
}
