package call

import (
	"math/rand"
	"time"
)

type Rand struct {
	rnd *rand.Rand
}

var current = time.Now

func NewRand() *Rand {
	r := rand.New(rand.NewSource(current().Unix()))
	return &Rand{
		rnd: r,
	}
}

func (r *Rand) bool() bool {
	r.rnd.Seed(current().Unix())
	return r.rnd.Intn(2)%2 == 0
}

func (r *Rand) int32() int32 {
	r.rnd.Seed(current().Unix())
	return r.rnd.Int31()
}

func (r *Rand) int64() int64 {
	r.rnd.Seed(current().Unix())
	return r.rnd.Int63()
}

func (r *Rand) uint32() uint32 {
	r.rnd.Seed(current().Unix())
	return r.rnd.Uint32()
}

func (r *Rand) uint64() uint64 {
	r.rnd.Seed(current().Unix())
	return r.rnd.Uint64()
}

func (r *Rand) float() float32 {
	r.rnd.Seed(current().Unix())
	return r.rnd.Float32()
}

func (r *Rand) double() float64 {
	r.rnd.Seed(current().Unix())
	return r.rnd.Float64()
}

func (r *Rand) pickupNum(len int) int {
	r.rnd.Seed(current().Unix())
	return r.rnd.Intn(len)
}

const defaultLength = 20
const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func (r *Rand) bytes() []byte {
	b := make([]byte, defaultLength)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return b
}

func (r *Rand) string() string {
	return string(r.bytes())
}
