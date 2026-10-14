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
	"reflect"
	"testing"
)

func TestGetWebAuthnRelyingParty(t *testing.T) {
	scenarios := []struct {
		description   string
		originBackend string
		rpId          string
		origins       string
		expectedRpId  string
		expectedList  []string
		expectError   bool
	}{
		{"defaults to backend origin", "https://id.example.com", "", "", "id.example.com", []string{"https://id.example.com"}, false},
		{"strips port from default RP ID", "http://localhost:8000", "", "", "localhost", []string{"http://localhost:8000"}, false},
		{"parent RP ID with extra origins", "https://id.example.com", "example.com", " https://console.example.com/, https://app.example.com ,", "example.com", []string{"https://id.example.com", "https://console.example.com", "https://app.example.com"}, false},
		{"deduplicates backend origin", "https://id.example.com", "Example.com", "https://id.example.com", "example.com", []string{"https://id.example.com"}, false},
		{"rejects origin outside RP ID", "https://id.example.com", "example.com", "https://console.other.com", "", nil, true},
		{"rejects backend outside RP ID", "https://id.example.com", "console.example.com", "https://console.example.com", "", nil, true},
		{"rejects suffix that is not a label boundary", "https://id.example.com", "ample.com", "", "", nil, true},
		{"rejects origin without scheme", "https://id.example.com", "example.com", "console.example.com", "", nil, true},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.description, func(t *testing.T) {
			rpId, origins, err := getWebAuthnRelyingParty(scenario.originBackend, scenario.rpId, scenario.origins)
			if scenario.expectError {
				if err == nil {
					t.Fatalf("expected error, got rpId=%s origins=%v", rpId, origins)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if rpId != scenario.expectedRpId {
				t.Errorf("rpId = %s, want %s", rpId, scenario.expectedRpId)
			}
			if !reflect.DeepEqual(origins, scenario.expectedList) {
				t.Errorf("origins = %v, want %v", origins, scenario.expectedList)
			}
		})
	}
}
