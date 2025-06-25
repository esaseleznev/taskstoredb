package contract

type EventType string

const (
	SetType    EventType = "set"
	DeleteType EventType = "del"
)

type Event struct {
	Type  EventType
	Key   []byte
	Value []byte
}
