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
)

// TestAppRecordTriggersTargetWebhookOnlyWhenAllowed — a user management call made with client
// credentials fires the provisioning webhook of the target organization only when the credentials
// belong to a "built-in" application and the call succeeded.
func TestAppRecordTriggersTargetWebhookOnlyWhenAllowed(t *testing.T) {
	initPasswordGrantTestDb(t)

	for _, application := range []*Application{
		{Owner: "admin", Name: "provisioner", Organization: "built-in"},
		{Owner: "admin", Name: "tenant-app", Organization: "other"},
	} {
		if _, err := ormer.Engine.Insert(application); err != nil {
			t.Fatal(err)
		}
	}
	webhook := &Webhook{Owner: "admin", Name: "acme-provisioning", Organization: "acme", Url: "https://example.com/hook", Method: "POST", ContentType: "application/json", Events: []string{"add-user"}, SingleOrgOnly: true, IsEnabled: true}
	if _, err := ormer.Engine.Insert(webhook); err != nil {
		t.Fatal(err)
	}

	countEvents := func() int64 {
		count, err := ormer.Engine.Count(&WebhookEvent{Webhook: webhook.GetId()})
		if err != nil {
			t.Fatal(err)
		}
		return count
	}

	newRecord := func(user string, organization string, response string) *Record {
		return &Record{
			Name:         "record-" + user + "-" + response[9:14],
			Organization: organization,
			User:         user,
			Method:       "POST",
			Action:       "add-user",
			Object:       `{"owner":"acme","name":"alice"}`,
			Response:     response,
		}
	}

	testCases := []struct {
		name         string
		record       *Record
		wantOrg      string
		wantNewEvent bool
	}{
		{"refused call of a built-in app", newRecord("app/provisioner", "built-in", `{status:"error", msg:"Unauthorized operation"}`), "built-in", false},
		{"call of another organization's app", newRecord("app/tenant-app", "other", `{status:"ok", msg:""}`), "other", false},
		{"unknown app", newRecord("app/missing", "app", `{status:"ok", msg:""}`), "app", false},
		{"successful call of a built-in app", newRecord("app/provisioner", "built-in", `{status:"ok", msg:""}`), "acme", true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			before := countEvents()
			AddRecord(testCase.record)
			if testCase.record.Organization != testCase.wantOrg {
				t.Fatalf("record organization = %q, want %q", testCase.record.Organization, testCase.wantOrg)
			}
			if gotNewEvent := countEvents() > before; gotNewEvent != testCase.wantNewEvent {
				t.Fatalf("webhook of the target organization fired = %v, want %v", gotNewEvent, testCase.wantNewEvent)
			}
		})
	}
}
