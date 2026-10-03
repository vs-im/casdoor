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

package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beego/beego/v2/server/web/context"
	"github.com/casdoor/casdoor/object"
)

func TestGetCapturedMagicLinkIs404WithoutFlag(t *testing.T) {
	t.Setenv(object.DevMagicLinkCaptureEnv, "")

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/get-captured-magic-link", strings.NewReader(`{"email":"e2e@x.test","organization":"built-in"}`))
	ctx := context.NewContext()
	ctx.Reset(recorder, request)
	ctx.Input.RequestBody = []byte(`{"email":"e2e@x.test","organization":"built-in"}`)

	c := &ApiController{}
	c.Ctx = ctx
	c.GetCapturedMagicLink()

	if ctx.Output.Status != http.StatusNotFound {
		t.Fatalf("want 404 without the flag, got %d", ctx.Output.Status)
	}
}
