package adapters

import "time"

const (
	prefixTask   = "t"
	prefixError  = "e"
	prefixGroup  = "g"
	prefixOwner  = "o"
	prefixOffset = "f"
)

type Offset struct {
	Value uint64    `json:"v"`
	Ts    time.Time `json:"t"`
}
