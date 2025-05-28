package common

import "sync/atomic"

type RoundRobin struct {
	owners []string
	num    uint32
}

func NewRoundRobind(owners ...string) *RoundRobin {
	return &RoundRobin{
		owners: owners,
	}
}

func (r *RoundRobin) Get() *string {
	len := len(r.owners)
	if len == 0 {
		return nil
	}
	num := int(atomic.AddUint32(&r.num, 1))
	return &r.owners[num%len]
}
