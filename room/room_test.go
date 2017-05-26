package room

import "testing"

func TestNormalize(t *testing.T) {
	testcases := []struct {
		desc string
		in   string
		want string
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
