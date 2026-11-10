// Copyright 2026 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package idp

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/markbates/goth"
)

func newAppleTestProvider(t *testing.T) *GothIdProvider {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	pemKey := string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
	p, err := NewGothIdProvider("Apple", "", "app.example.ios", "TEAMID1234", "KEYID12345", pemKey, "https://example.com/callback", "")
	if err != nil {
		t.Fatalf("apple provider: %v", err)
	}
	return p
}

type appleTokenCall struct {
	authorization string
	form          url.Values
}

// The code exchange is one POST with the credentials in the body and no Authorization header, and a refusal
// is returned as it is, not replaced by a second attempt with the same code.
func TestAppleExchangeIsSinglePost(t *testing.T) {
	var mu sync.Mutex
	var calls []appleTokenCall
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		mu.Lock()
		calls = append(calls, appleTokenCall{authorization: r.Header.Get("Authorization"), form: r.PostForm})
		n := len(calls)
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if n == 1 {
			_, _ = w.Write([]byte(`{"error":"invalid_client","error_description":"client_secret is invalid"}`))
			return
		}
		_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
	}))
	defer srv.Close()

	p := newAppleTestProvider(t)
	cfg, err := appleOAuth2Config(p.Provider)
	if err != nil {
		t.Fatal(err)
	}
	cfg.Endpoint.TokenURL = srv.URL

	_, err = p.Session.Authorize(p.Provider, goth.Params(url.Values{"code": {"one-time-code"}}))
	if err == nil {
		t.Fatal("a refused exchange must fail")
	}
	if !strings.Contains(err.Error(), "invalid_client") {
		t.Fatalf("the answer to the first POST must reach the caller, got %v", err)
	}
	if len(calls) != 1 {
		t.Fatalf("the token endpoint got %d requests, want 1", len(calls))
	}
	c := calls[0]
	if c.authorization != "" {
		t.Fatalf("Authorization header must not be sent, got %q", c.authorization)
	}
	for k, want := range map[string]string{"code": "one-time-code", "grant_type": "authorization_code", "client_id": "app.example.ios"} {
		if got := c.form.Get(k); got != want {
			t.Fatalf("form %s = %q, want %q", k, got, want)
		}
	}
	if c.form.Get("client_secret") == "" || len(c.form["client_id"]) != 1 || len(c.form["client_secret"]) != 1 {
		t.Fatalf("client_id and client_secret must be sent once in the body: %v", c.form)
	}
}
