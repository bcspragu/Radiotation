package db

import (
	"math/rand"
	"testing"
)

func TestNormalize(t *testing.T) {
	testcases := []struct {
		desc string
		in   string
		want RoomID
	}{
		{"empty name should return blank", "", "blank"},
		{"long name should be truncated", "thisnameiswaytoolong", "thisnameiswayto"},
		{"capital letters should be lower-cased", "YeLLiNG", "yelling"},
		{"dashes and spaces should become hyphens", "what_s goin-on", "what-s-goin-on"},
		{"other non-alphanumerics should be removed", "!@#$te st%^&*", "te-st"},
		{"more non-alphanumerics should be removed", "(){}te st:\"<>?", "te-st"},
	}

	for _, tc := range testcases {
		got := Normalize(tc.in)
		if got != tc.want {
			t.Errorf("%s: Normalize(%s) = %s, want %s", tc.desc, tc.in, got, tc.want)
		}
	}
}

func TestRoundRobinRotator_Empty(t *testing.T) {
	r := &roundRobinRotator{}

	r.Start(0)
	tryNext(t, r, 0, true)

	r.Start(0)
	tryNext(t, r, 0, true)
}

func TestRoundRobinRotator(t *testing.T) {
	r := &roundRobinRotator{}

	r.Start(1)
	tryNext(t, r, 0, true)

	r.Start(2)
	tryNext(t, r, 0, false)
	tryNext(t, r, 1, true)

	r.Start(3)
	tryNext(t, r, 0, false)
	tryNext(t, r, 1, false)
	tryNext(t, r, 2, true)

	r.Start(4)
	tryNext(t, r, 0, false)
	tryNext(t, r, 1, false)
	// Restart in middle
	r.Start(4)
	tryNext(t, r, 0, false)
	tryNext(t, r, 1, false)
	tryNext(t, r, 2, false)
	tryNext(t, r, 3, true)
}

func TestShuffleRotator_Empty(t *testing.T) {
	r := &shuffleRotator{R: rand.New(rand.NewSource(0))}

	r.Start(0)
	tryNext(t, r, 0, true)

	r.Start(0)
	tryNext(t, r, 0, true)
}

func TestShuffleRotator(t *testing.T) {
	r := &shuffleRotator{R: rand.New(rand.NewSource(0))}
	r.Start(1)
	tryNext(t, r, 0, true)

	r = &shuffleRotator{R: rand.New(rand.NewSource(0))}
	r.Start(2)
	tryNext(t, r, 1, false)
	tryNext(t, r, 0, true)

	r = &shuffleRotator{R: rand.New(rand.NewSource(0))}
	r.Start(3)
	tryNext(t, r, 1, false)
	tryNext(t, r, 2, false)
	tryNext(t, r, 0, true)

	r = &shuffleRotator{R: rand.New(rand.NewSource(3))}
	r.Start(4)
	tryNext(t, r, 2, false)
	tryNext(t, r, 1, false)
	// Restart in middle
	r = &shuffleRotator{R: rand.New(rand.NewSource(3))}
	r.Start(4)
	tryNext(t, r, 2, false)
	tryNext(t, r, 1, false)
	tryNext(t, r, 3, false)
	tryNext(t, r, 0, true)
}

func TestRandomRotator_Empty(t *testing.T) {
	r := &randomRotator{R: rand.New(rand.NewSource(0))}

	r.Start(0)
	tryNext(t, r, 0, true)

	r.Start(0)
	tryNext(t, r, 0, true)
}

func TestRandomRotator(t *testing.T) {
	r := &randomRotator{R: rand.New(rand.NewSource(0))}

	for i := 0; i < 10; i++ {
		r.Start(1)
		tryNext(t, r, 0, true)
	}

	r = &randomRotator{R: rand.New(rand.NewSource(0))}
	r.Start(2)
	tryNext(t, r, 0, true)
	tryNext(t, r, 0, true)
	r.Start(2)
	tryNext(t, r, 1, true)
	tryNext(t, r, 0, true)

	r = &randomRotator{R: rand.New(rand.NewSource(0))}
	r.Start(3)
	tryNext(t, r, 0, true)
	tryNext(t, r, 0, true)
	tryNext(t, r, 1, true)
	r.Start(3)
	tryNext(t, r, 1, true)
	tryNext(t, r, 2, true)
	tryNext(t, r, 1, true)

	r = &randomRotator{R: rand.New(rand.NewSource(3))}
	r.Start(4)
	tryNext(t, r, 0, true)
	tryNext(t, r, 1, true)
	// Restart in middle
	r = &randomRotator{R: rand.New(rand.NewSource(3))}
	r.Start(4)
	tryNext(t, r, 0, true)
	tryNext(t, r, 1, true)
	tryNext(t, r, 0, true)
	tryNext(t, r, 2, true)
}

func tryNext(t *testing.T, r Rotator, idx int, last bool) {
	if i, l := r.NextIndex(); i != idx || l != last {
		t.Errorf("r.NextIndex() = (%d, %t), want (%d, %t)", i, l, idx, false)
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
