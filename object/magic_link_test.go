package object

import "testing"

func TestMagicLinkTokenHash(t *testing.T) {
	token, err := GenerateMagicLinkToken()
	if err != nil {
		t.Fatal(err)
	}
	if token == "" {
		t.Fatal("token should not be empty")
	}

	hash := HashMagicLinkToken(token)
	if hash == "" {
		t.Fatal("hash should not be empty")
	}
	if hash == token {
		t.Fatal("hash should not equal token")
	}
	if hash != HashMagicLinkToken(token) {
		t.Fatal("hash should be stable")
	}
}
