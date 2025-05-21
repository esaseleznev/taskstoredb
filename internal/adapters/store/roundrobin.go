package adapters

import "sync/atomic"

type roundRobin struct {
	owners []string
	num    uint32
}

func newRoundRobind(owners ...string) *roundRobin {
	return &roundRobin{
		owners: owners,
	}
}

func (r *roundRobin) get() *string {
	len := len(r.owners)
	if len == 0 {
		return nil
	}
	num := int(atomic.AddUint32(&r.num, 1))
	return &r.owners[num%len]
}
