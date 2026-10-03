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
	"strings"
	"sync"
	"time"

	"github.com/casdoor/casdoor/conf"
)

// Dev capture of the magic links for end-to-end tests. With CASDOOR_DEV_MAGIC_LINK_CAPTURE=true
// (off by default) a link of the API addressed to a ".test" mailbox is not mailed: its open
// link is kept in memory, the last one per address, until the TTL runs out, and a server
// holding the application's credentials reads it from /api/get-captured-magic-link.

const (
	DevMagicLinkCaptureEnv = "CASDOOR_DEV_MAGIC_LINK_CAPTURE"

	devMagicLinkCaptureTTL = 10 * time.Minute
	devMagicLinkCaptureMax = 1000
)

type capturedMagicLink struct {
	link      string
	expiresAt time.Time
}

var (
	capturedMagicLinks   = map[string]capturedMagicLink{}
	capturedMagicLinksMu sync.Mutex
)

func IsDevMagicLinkCaptureEnabled() bool {
	return conf.GetConfigBool(DevMagicLinkCaptureEnv)
}

// ShouldCaptureMagicLink tells whether the link to this address is captured instead of mailed.
func ShouldCaptureMagicLink(email string) bool {
	return IsDevMagicLinkCaptureEnabled() && isDevTestMailbox(email)
}

func isDevTestMailbox(email string) bool {
	email = strings.ToLower(strings.TrimSpace(email))
	at := strings.LastIndex(email, "@")
	return at > 0 && strings.HasSuffix(email, ".test") && len(email) > at+len(".test")+1
}

func captureMagicLink(applicationId string, email string, link string, now time.Time) {
	capturedMagicLinksMu.Lock()
	defer capturedMagicLinksMu.Unlock()

	for key, entry := range capturedMagicLinks {
		if !entry.expiresAt.After(now) {
			delete(capturedMagicLinks, key)
		}
	}
	key := capturedMagicLinkKey(applicationId, email)
	if _, ok := capturedMagicLinks[key]; !ok && len(capturedMagicLinks) >= devMagicLinkCaptureMax {
		return
	}
	capturedMagicLinks[key] = capturedMagicLink{link: link, expiresAt: now.Add(devMagicLinkCaptureTTL)}
}

// GetCapturedMagicLink returns the last captured link of the address, if it has not expired.
func GetCapturedMagicLink(applicationId string, email string) (string, bool) {
	return getCapturedMagicLinkAt(applicationId, email, time.Now())
}

func getCapturedMagicLinkAt(applicationId string, email string, now time.Time) (string, bool) {
	capturedMagicLinksMu.Lock()
	defer capturedMagicLinksMu.Unlock()

	entry, ok := capturedMagicLinks[capturedMagicLinkKey(applicationId, email)]
	if !ok || !entry.expiresAt.After(now) {
		return "", false
	}
	return entry.link, true
}

func capturedMagicLinkKey(applicationId string, email string) string {
	return applicationId + "|" + strings.ToLower(strings.TrimSpace(email))
}
