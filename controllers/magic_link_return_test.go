package controllers

import "testing"

func TestReturnLinkAllowedOnlyForTrustedSend(t *testing.T) {
	cases := []struct {
		name    string
		form    MagicLinkRequestForm
		trusted bool
		want    bool
	}{
		{"plain request", MagicLinkRequestForm{}, false, true},
		{"plain trusted request", MagicLinkRequestForm{Trusted: true}, true, true},
		{"returnLink untrusted", MagicLinkRequestForm{ReturnLink: true}, false, false},
		{"returnLink trusted", MagicLinkRequestForm{ReturnLink: true, Trusted: true}, true, true},
	}
	for _, tc := range cases {
		if got := returnLinkAllowed(&tc.form, tc.trusted); got != tc.want {
			t.Fatalf("%s: got %v, want %v", tc.name, got, tc.want)
		}
	}
}
