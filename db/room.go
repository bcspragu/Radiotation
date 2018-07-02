package db

import (
	"bytes"
	"encoding/gob"
	"log"
	"math/rand"
	"time"

	"github.com/bcspragu/Radiotation/rng"
)

type (
	// RotatorType is an enum for the type of rotation for the room.
	RotatorType int

	// RoomID is an opaque string identifying a room.
	RoomID string

	Room struct {
		ID          RoomID
		DisplayName string
		RotatorType RotatorType
	}
)

func init() {
	gob.Register(&roundRobinRotator{})
	gob.Register(&shuffleRotator{})
	gob.Register(&randomRotator{})
}

const (
	RoundRobin RotatorType = iota
	Shuffle
	Random
)

func (r RotatorType) String() string {
	switch r {
	case RoundRobin:
		return "Round Robin"
	case Shuffle:
		return "Shuffle"
	case Random:
		return "Random"
	}
	return "Unknown"
}

func NewRotator(r RotatorType) Rotator {
	switch r {
	case RoundRobin:
		return newRoundRobin()
	case Shuffle:
		return newShuffle(rng.NewSource(time.Now().Unix()))
	case Random:
		return newRandom(rng.NewSource(time.Now().Unix()))
	default:
		return nil
	}
}

func newRoundRobin() *roundRobinRotator {
	return &roundRobinRotator{}
}

func newShuffle(src *rng.Source) *shuffleRotator {
	return &shuffleRotator{Src: src, AddNext: make(map[int]int)}
}

func newRandom(src *rng.Source) *randomRotator {
	return &randomRotator{Src: src}
}

// A Rotator says what the next index should be from a list of size n. They can
// assume that n will never decrease, but it can increase between any two
// invocations of start.
type Rotator interface {
	// NextIndex returns the next index in the rotation.
	NextIndex() int
	// Add adds a new index into the rotation. Depending on the implementation,
	// this may or may not be at the end. It's invalid to call NextIndex() before
	// Add() has been called at least once.
	Add()
}

type randomRotator struct {
	N   int
	Src *rng.Source
}

func LoadRotator(b []byte) Rotator {
	// The decode will fail unless the concrete type on the wire has been
	// registered. We registered it in the calling function.
	var r Rotator
	err := gob.NewDecoder(bytes.NewReader(b)).Decode(&r)
	if err != nil {
		log.Printf("decode rotator: %v", err)
	}
	return r
}

func SaveRotator(r Rotator) []byte {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(&r)
	if err != nil {
		log.Printf("encode rotator: %v", err)
	}
	return buf.Bytes()
}

func (r *randomRotator) NextIndex() int {
	if r.N == 0 {
		return 0
	}
	return rand.New(r.Src).Intn(r.N)
}

func (r *randomRotator) Add() {
	r.N++
}

type shuffleRotator struct {
	Perm  []int
	Index int
	// AddNext is a map from a user to the location in the queue to add them on
	// our next rotation, to make sure everyone gets a song not too long after
	// joining.
	AddNext map[int]int
	Src     *rng.Source
}

func (s *shuffleRotator) NextIndex() int {
	if len(s.Perm) == 0 {
		return 0
	}

	if s.Index >= len(s.Perm) {
		s.newRotation()
	}

	i := s.Perm[s.Index]
	s.Index++
	return i
}

func (s *shuffleRotator) Add() {
	// Insert half way into the queue.
	i := (s.Index + (len(s.Perm)+1)/2) % (len(s.Perm) + 1)

	// If we've already passed the index, add them for next time.
	if i < s.Index {
		s.AddNext[len(s.Perm)] = i
		return
	}

	// Otherwise, just add them into the queue.
	s.Perm = append(s.Perm, 0)
	copy(s.Perm[i+1:], s.Perm[i:])
	s.Perm[i] = len(s.Perm) - 1
}

func (s *shuffleRotator) newRotation() {
	s.Perm = rand.New(s.Src).Perm(len(s.Perm))

	for v, i := range s.AddNext {
		s.Perm = append(s.Perm, 0)
		copy(s.Perm[i+1:], s.Perm[i:])
		s.Perm[i] = v

		delete(s.AddNext, v)
	}

	s.Index = 0
}

// roundRobinRotator goes through users in the same order every time. When a
// new user is added to the rotation, they are added in the middle.
type roundRobinRotator struct {
	Order []int
	Index int
}

func (r *roundRobinRotator) NextIndex() int {
	if len(r.Order) == 0 {
		return 0
	}

	i := r.Order[r.Index]
	r.Index = (r.Index + 1) % len(r.Order)
	return i
}

func (r *roundRobinRotator) Add() {
	// Insert half way into the queue.
	i := (r.Index + (len(r.Order)+1)/2) % (len(r.Order) + 1)

	r.Order = append(r.Order, 0)
	copy(r.Order[i+1:], r.Order[i:])
	r.Order[i] = len(r.Order) - 1
}
