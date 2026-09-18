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

package object

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beego/beego/v2/server/web/context"
)

func newMagicLinkRecordContext(method string, uri string, body string) *context.Context {
	ctx := context.NewContext()
	ctx.Reset(httptest.NewRecorder(), httptest.NewRequest(method, uri, strings.NewReader(body)))
	ctx.Request.RequestURI = uri
	ctx.Input.RequestBody = []byte(body)
	return ctx
}

func TestRecordKeepsMagicLinkTokenOut(t *testing.T) {
	record, err := NewRecord(newMagicLinkRecordContext("GET", "/api/verify-magic-link?token=raw-token&sessionSecret=raw-secret&magicLinkToken=raw-built-in&client_id=abc", ""))
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"raw-token", "raw-secret", "raw-built-in"} {
		if strings.Contains(record.RequestUri, secret) {
			t.Fatalf("the record uri keeps %q: %s", secret, record.RequestUri)
		}
	}
	if !strings.Contains(record.RequestUri, "client_id=abc") {
		t.Fatalf("the rest of the query should stay: %s", record.RequestUri)
	}

	record, err = NewRecord(newMagicLinkRecordContext("POST", "/api/login", `{"application":"app","signinMethod":"Magic link","code":"raw-token"}`))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(record.Object, "raw-token") || !strings.Contains(record.Object, `"application":"app"`) {
		t.Fatalf("the record object keeps the token of the built-in sign-in: %s", record.Object)
	}

	record, err = NewRecord(newMagicLinkRecordContext("POST", "/api/login", `{"application":"app","signinMethod":"Verification code","code":"123456"}`))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(record.Object, "123456") {
		t.Fatalf("other sign-in methods are recorded as before: %s", record.Object)
	}

	record, err = NewRecord(newMagicLinkRecordContext("POST", "/api/send-magic-link", `{"email":"a@example.com","sessionSecret":"raw-secret"}`))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(record.Object, "raw-secret") {
		t.Fatalf("the record object keeps the session secret: %s", record.Object)
	}
}
