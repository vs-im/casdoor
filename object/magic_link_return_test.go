package object

import (
	"strings"
	"testing"
	"time"
)

// A returned link is stored, usable once from any browser, and carries the caller's OAuth request.
func TestIssueApiMagicLinkReturnsUsableOneTimeLink(t *testing.T) {
	a := newMagicLinkTestOrmer(t)
	if err := a.syncMagicLink(); err != nil {
		t.Fatal(err)
	}
	application := newMagicLinkTestApplication()
	oauth := map[string]string{"redirectUri": "https://app.example.com/cb"}
	token, link, err := NewApiMagicLink(application, nil, "a@example.com", "192.0.2.1", "", oauth, time.Time{})
	if err != nil {
		t.Fatal(err)
	}

	callbackURL, err := issueApiMagicLinkAt(link, token, "https://id.example.com/")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(callbackURL, "https://id.example.com/magic-link/callback?") || !strings.Contains(callbackURL, "token="+token) {
		t.Fatalf("unexpected link: %s", callbackURL)
	}
	if !strings.Contains(callbackURL, "redirect_uri=https%3A%2F%2Fapp.example.com%2Fcb") {
		t.Fatalf("the OAuth request is lost: %s", callbackURL)
	}

	stored, err := GetMagicLinkByToken(token)
	if err != nil || stored == nil || stored.Status != MagicLinkStatusSent {
		t.Fatalf("the link must be stored as sent: %+v %v", stored, err)
	}
	if _, err = ConsumeMagicLink(token, getApiMagicLinkSessionHash(stored, "", ""), application, "en"); err != nil {
		t.Fatalf("an unbound link opens on any device: %v", err)
	}
	if _, err = ConsumeMagicLink(token, getApiMagicLinkSessionHash(stored, "", ""), application, "en"); err == nil {
		t.Fatal("the link must work once")
	}
}
