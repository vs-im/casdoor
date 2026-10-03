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
	"testing"
	"time"
)

func TestIsDevTestMailbox(t *testing.T) {
	for email, want := range map[string]bool{
		"e2e-1@example.test": true,
		" E2E@Mail.TEST ":    true,
		"a@b.test":           true,
		"a@example.com":      false,
		"a@test":             false,
		"a@.test":            false,
		"@x.test":            false,
		"x.test":             false,
		"":                   false,
	} {
		if got := isDevTestMailbox(email); got != want {
			t.Errorf("isDevTestMailbox(%q) = %v, want %v", email, got, want)
		}
	}
}

func TestShouldCaptureMagicLinkNeedsFlag(t *testing.T) {
	t.Setenv(DevMagicLinkCaptureEnv, "")
	if ShouldCaptureMagicLink("e2e-1@example.test") {
		t.Fatal("capture must be off by default")
	}
	t.Setenv(DevMagicLinkCaptureEnv, "true")
	if !ShouldCaptureMagicLink("e2e-1@example.test") {
		t.Fatal("capture must be on with the flag")
	}
	if ShouldCaptureMagicLink("user@example.com") {
		t.Fatal("a real address must never be captured")
	}
}

func TestCapturedMagicLinkKeepsLastWithTTL(t *testing.T) {
	now := time.Now()
	app := "org/app-ttl"
	captureMagicLink(app, "E2E-ttl@x.test", "https://a/1", now)
	captureMagicLink(app, "e2e-ttl@x.test", "https://a/2", now)
	if got, ok := getCapturedMagicLinkAt(app, "e2e-ttl@x.test", now.Add(time.Minute)); !ok || got != "https://a/2" {
		t.Fatalf("want the last link, got %q %v", got, ok)
	}
	if _, ok := getCapturedMagicLinkAt("org/other", "e2e-ttl@x.test", now); ok {
		t.Fatal("links of another application must not leak")
	}
	if _, ok := getCapturedMagicLinkAt(app, "e2e-ttl@x.test", now.Add(devMagicLinkCaptureTTL+time.Second)); ok {
		t.Fatal("an expired link must not be returned")
	}
	captureMagicLink(app, "e2e-other@x.test", "https://a/3", now.Add(devMagicLinkCaptureTTL+time.Minute))
	capturedMagicLinksMu.Lock()
	_, stale := capturedMagicLinks[capturedMagicLinkKey(app, "e2e-ttl@x.test")]
	capturedMagicLinksMu.Unlock()
	if stale {
		t.Fatal("expired links must be purged on write")
	}
}
