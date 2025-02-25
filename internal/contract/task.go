package contract

import "time"

type Task struct {
	Id     string            `json:"-"`
	Kind   string            `json:"k"`
	Group  string            `json:"g"`
	Owner  *string           `json:"o"`
	Status Status            `json:"s"`
	Param  map[string]string `json:"p"`
	Ts     time.Time         `json:"t"`
	Error  *string           `json:"e"`
}

type Status int

const (
	VIRGIN    Status = 1
	SCHEDULED Status = 2
	COMPLETED Status = 3
	FAILED    Status = 4
)
