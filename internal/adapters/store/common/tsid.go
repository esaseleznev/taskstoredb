package common

import "sync"

const (
	epoch         int64 = 1735678800000 // 2025-01-01T00:00:00.000Z
	randomBitSize int64 = 22
)

var alphabets []rune = []rune("0123456789ABCDEFGHJKMNPQRSTVWXYZ")

type Tsid struct {
	mu  *sync.Mutex
	ts  int64
	num int32
}

func NewTsid() *Tsid {
	return &Tsid{mu: &sync.Mutex{}}
}

func (l *Tsid) Next(ts int64) string {
	l.mu.Lock()
	defer l.mu.Unlock()
	ts = (ts - epoch) << randomBitSize
	if ts == l.ts {
		l.num += 1
	} else {
		l.num = 1
		l.ts = ts
	}

	val := ts | int64(l.num)
	return l.toString(val)
}

func (t *Tsid) toString(number int64) string {
	chars := make([]rune, 13)

	chars[0] = alphabets[((number >> 60) & 0b11111)]
	chars[1] = alphabets[((number >> 55) & 0b11111)]
	chars[2] = alphabets[((number >> 50) & 0b11111)]
	chars[3] = alphabets[((number >> 45) & 0b11111)]
	chars[4] = alphabets[((number >> 40) & 0b11111)]
	chars[5] = alphabets[((number >> 35) & 0b11111)]
	chars[6] = alphabets[((number >> 30) & 0b11111)]
	chars[7] = alphabets[((number >> 25) & 0b11111)]
	chars[8] = alphabets[((number >> 20) & 0b11111)]
	chars[9] = alphabets[((number >> 15) & 0b11111)]
	chars[10] = alphabets[((number >> 10) & 0b11111)]
	chars[11] = alphabets[((number >> 5) & 0b11111)]
	chars[12] = alphabets[(number & 0b11111)]

	return string(chars)
}
