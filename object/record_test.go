package object

import (
	"strings"
	"testing"
)

func TestMaskSensitiveFields(t *testing.T) {
	body := `{
		"password": "pass",
		"clientSecret" : "secret",
		"client_secret": "snake",
		"applicationClientSecret" : "app-secret",
		"nested": [{"ClientSecret": "nested-secret"}],
		"email": "user@example.com"
	}`
	masked := maskSensitiveFields(body)
	for _, value := range []string{`"pass"`, `"secret"`, `"snake"`, `"app-secret"`, `"nested-secret"`} {
		if strings.Contains(masked, value) {
			t.Fatalf("expected %s to be masked in %s", value, masked)
		}
	}
	for _, value := range []string{`"password":"***"`, `"clientSecret":"***"`, `"client_secret":"***"`, `"applicationClientSecret":"***"`} {
		if !strings.Contains(masked, value) {
			t.Fatalf("expected %s in %s", value, masked)
		}
	}
}

func TestGetRecordTargetOrganization(t *testing.T) {
	testCases := []struct {
		name     string
		action   string
		object   string
		expected string
	}{
		{"add-user resolves owner", "add-user", `{"owner":"acme","name":"alice"}`, "acme"},
		{"update-user resolves owner", "update-user", `{"owner":"acme","name":"alice"}`, "acme"},
		{"delete-user resolves owner", "delete-user", `{"owner":"acme","name":"alice"}`, "acme"},
		{"add-organization resolves name", "add-organization", `{"owner":"admin","name":"acme"}`, "acme"},
		{"update-organization resolves name", "update-organization", `{"owner":"admin","name":"acme"}`, "acme"},
		{"delete-organization resolves name", "delete-organization", `{"owner":"admin","name":"acme"}`, "acme"},
		{"non-provisioning action is dropped", "get-users", `{"owner":"acme"}`, ""},
		{"login action is dropped", "login", `{"owner":"acme"}`, ""},
		{"invalid JSON is dropped", "add-user", `{owner:`, ""},
		{"missing field is dropped", "add-user", `{"name":"alice"}`, ""},
		{"non-string field is dropped", "add-user", `{"owner":42}`, ""},
		{"empty object is dropped", "add-user", ``, ""},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			record := &Record{Organization: "app", Action: testCase.action, Object: testCase.object}
			result := getRecordTargetOrganization(record)
			if result != testCase.expected {
				t.Fatalf("action %s object %s: expected %q, got %q", testCase.action, testCase.object, testCase.expected, result)
			}
		})
	}
}

func TestMaskSensitiveFieldsFallback(t *testing.T) {
	body := `{"clientSecret" : "secret", "password" : "pass"`
	masked := maskSensitiveFields(body)
	if strings.Contains(masked, `"secret"`) || strings.Contains(masked, `"pass"`) {
		t.Fatalf("expected fallback redaction in %s", masked)
	}
}
