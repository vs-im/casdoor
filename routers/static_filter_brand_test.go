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

package routers

import (
	"os"
	"strings"
	"testing"

	"github.com/casdoor/casdoor/conf"
)

const indexHtmlPath = "../web/index.html"

func TestApplyBrandToIndexHtmlKeepsUpstreamByDefault(t *testing.T) {
	shell, err := os.ReadFile(indexHtmlPath)
	if err != nil {
		t.Skipf("no frontend shell to check: %v", err)
	}

	got := applyBrandToIndexHtml(string(shell))

	if !strings.Contains(got, "<title>"+conf.DefaultBrandName+"</title>") {
		t.Error("an unbranded deployment must keep the upstream title")
	}
	// The repository ships no favicon of its own, so the unbranded shell points at
	// the upstream CDN one.
	if !strings.Contains(got, conf.DefaultBrandFaviconUrl) {
		t.Errorf("an unbranded deployment must fall back to the upstream favicon %q", conf.DefaultBrandFaviconUrl)
	}
}

func TestApplyBrandToIndexHtml(t *testing.T) {
	shell, err := os.ReadFile(indexHtmlPath)
	if err != nil {
		t.Skipf("no frontend shell to check: %v", err)
	}

	t.Setenv(conf.BrandNameEnv, "Acme Identity")
	t.Setenv(conf.BrandFaviconUrlEnv, "/brand/favicon.png")

	got := applyBrandToIndexHtml(string(shell))

	for _, expected := range []string{
		"<title>Acme Identity</title>",
		`content="Acme Identity - sign in"`,
		`href="/brand/favicon.png"`,
	} {
		if !strings.Contains(got, expected) {
			t.Errorf("the branded shell is missing %q", expected)
		}
	}
	if strings.Contains(got, "<title>Casdoor</title>") || strings.Contains(got, `href="/favicon.png"`) {
		t.Error("the branded shell still carries the upstream title or favicon")
	}
}
