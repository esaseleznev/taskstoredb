package common

import (
	"github.com/esaseleznev/taskstoredb/internal/contract"
)

type Playload struct {
	data []contract.Event
}

func NewPlayload() *Playload {
	return &Playload{data: []contract.Event{}}
}

func (p *Playload) Data() []contract.Event {
	return p.data
}

func (p *Playload) Put(key []byte, value []byte) {
	p.data = append(p.data, contract.Event{Key: key, Value: value, Type: contract.SetType})
}

func (p *Playload) Delete(key []byte, value []byte) {
	p.data = append(p.data, contract.Event{Key: key, Value: value, Type: contract.DeleteType})
}
