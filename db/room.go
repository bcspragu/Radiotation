package db

import (
	"bytes"
	"encoding/gob"
	"log"
	"math/rand"
	"regexp"
	"strings"
	"time"
)

var (
	nameRE = regexp.MustCompile(`[^a-zA-Z0-9\-]+`)
)

type (
	// MusicService is an enum for supported music streaming platforms.
	MusicService int
	// RotatorType is an enum for the type of rotation for the room.
	RotatorType int

	// RoomID is an opaque string identifying a room.
	RoomID string

	Room struct {
		ID           RoomID
		DisplayName  string
		Rotator      Rotator
		MusicService MusicService
	}
)

func init() {
	gob.Register(&roundRobinRotator{})
	gob.Register(&shuffleRotator{})
	gob.Register(&randomRotator{})
}

const (
	PlayMusic MusicService = iota
	Spotify
)

const (
	RoundRobin RotatorType = iota
	Shuffle
	Random
)

func (m MusicService) String() string {
	switch m {
	case PlayMusic:
		return "Play Music"
	case Spotify:
		return "Spotify"
	}
	return "Unknown"
}

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
		return &roundRobinRotator{}
	case Shuffle:
		return &shuffleRotator{R: rand.New(rand.NewSource(time.Now().Unix()))}
	case Random:
		return &randomRotator{R: rand.New(rand.NewSource(time.Now().Unix()))}
	default:
		return nil
	}
}

// A Rotator says what the next index should be from a list of size n. They can
// assume that n will never decrease, but it can increase between any two
// invocations of start.
type Rotator interface {
	// NextIndex returns the next index in the current rotation, and if this is
	// the last entry in the current rotation
	NextIndex() (int, bool)
	// Start takes the size of the current rotation and should be called before
	// we do our first rotation, and after each subsequent rotation is over
	Start(n int)
}

type randomRotator struct {
	N int
	R *rand.Rand
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

func (r *randomRotator) NextIndex() (int, bool) {
	if r.N == 0 {
		return 0, true
	}
	return r.R.Intn(r.N), true
}

func (r *randomRotator) Start(n int) {
	r.N = n
}

type shuffleRotator struct {
	ShuffleList []int
	Index       int
	R           *rand.Rand
}

func (s *shuffleRotator) NextIndex() (int, bool) {
	if len(s.ShuffleList) == 0 {
		return 0, true
	}

	if s.Index >= len(s.ShuffleList) {
		// Keep returning the last element if we've gone too far
		return s.ShuffleList[len(s.ShuffleList)-1], true
	}

	i := s.ShuffleList[s.Index]
	s.Index++
	return i, s.Index == len(s.ShuffleList)
}

func (s *shuffleRotator) Start(n int) {
	s.ShuffleList = s.R.Perm(n)
	s.Index = 0
}

type roundRobinRotator struct {
	Offset int
	N      int
}

func (r *roundRobinRotator) NextIndex() (int, bool) {
	if r.N == 0 {
		return 0, true
	}

	i := r.Offset
	r.Offset = (r.Offset + 1) % r.N
	return i, r.Offset == 0
}

func (r *roundRobinRotator) Start(n int) {
	r.Offset = 0
	r.N = n
}

// NewRoom initializes a new room with no users.
func NewRoom(name string, ms MusicService) *Room {
	return &Room{
		DisplayName:  name,
		ID:           Normalize(name),
		MusicService: ms,
	}
}

func Normalize(name string) RoomID {
	if len(name) == 0 {
		name = "blank"
	}

	if len(name) > 15 {
		name = name[:15]
	}
	name = strings.ToLower(name)
	name = strings.TrimSpace(name)
	name = strings.Replace(name, " ", "-", -1)
	name = strings.Replace(name, "_", "-", -1)
	return RoomID(nameRE.ReplaceAllString(name, ""))
}
