package srv

import (
	"testing"

	"github.com/bcspragu/Radiotation/db"
	"github.com/google/go-cmp/cmp"
)

func TestContinuationToken(t *testing.T) {
	ct := &continuationToken{
		HistoryIndex: 17,
		RoomID:       db.RoomID("room123"),
		UserID:       db.UserID("user123"),
		TrackID:      "track123",
	}

	str, err := makeContinuationToken(ct)
	if err != nil {
		t.Fatalf("makeContinuationToken: %v", err)
	}

	got, err := parseContinuationToken(str)
	if err != nil {
		t.Fatalf("parseContinuationToken: %v", err)
	}

	if diff := cmp.Diff(ct, got); diff != "" {
		t.Fatalf("got unexpected continuation token (-want +got):\n%s", diff)
	}
}
