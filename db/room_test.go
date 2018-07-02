package db

import (
	"testing"

	"github.com/bcspragu/Radiotation/rng"
)

func TestRoundRobinRotator(t *testing.T) {
	r := newRoundRobin()

	r.Add()
	tryNext(t, r, 0)

	r.Add()
	tryNext(t, r, 0)
	tryNext(t, r, 1)

	r.Add()
	tryNext(t, r, 0)
	tryNext(t, r, 2)
	tryNext(t, r, 1)

	r.Add()
	tryNext(t, r, 0)
	tryNext(t, r, 2)
	tryNext(t, r, 3)
	tryNext(t, r, 1)
	tryNext(t, r, 0)
	tryNext(t, r, 2)

	// Add someone in the middle
	r.Add()
	tryNext(t, r, 3)
	tryNext(t, r, 1)
	tryNext(t, r, 4)
	tryNext(t, r, 0)
	tryNext(t, r, 2)
	tryNext(t, r, 3)
	tryNext(t, r, 1)
}

func TestShuffleRotator(t *testing.T) {
	r := newShuffle(rng.NewSource(0))

	r.Add()
	tryNext(t, r, 0)

	r.Add()
	tryNext(t, r, 1)
	tryNext(t, r, 0)
	tryNext(t, r, 0)
	tryNext(t, r, 1)
	tryNext(t, r, 0)

	r.Add()
	tryNext(t, r, 1)
	tryNext(t, r, 2)
	tryNext(t, r, 0)
	tryNext(t, r, 1)
	tryNext(t, r, 2)
	tryNext(t, r, 2)
	tryNext(t, r, 0)

	r.Add()
	tryNext(t, r, 1)
	tryNext(t, r, 3)
}

func TestRandomRotator(t *testing.T) {
	r := newRandom(rng.NewSource(0))

	r.Add()
	tryNext(t, r, 0)
	tryNext(t, r, 0)

	r.Add()
	tryNext(t, r, 1)
	tryNext(t, r, 0)
	tryNext(t, r, 1)
	tryNext(t, r, 0)
	tryNext(t, r, 1)

	r.Add()
	tryNext(t, r, 2)
	tryNext(t, r, 0)
	tryNext(t, r, 0)
	tryNext(t, r, 0)
	tryNext(t, r, 2)
}

func tryNext(t *testing.T, r Rotator, wantIdx int) {
	t.Helper()
	if gotIdx := r.NextIndex(); wantIdx != gotIdx {
		t.Errorf("r.NextIndex() = (%d), want (%d)", gotIdx, wantIdx)
	}
}
