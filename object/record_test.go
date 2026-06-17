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

func TestMaskSensitiveFieldsFallback(t *testing.T) {
	body := `{"clientSecret" : "secret", "password" : "pass"`
	masked := maskSensitiveFields(body)
	if strings.Contains(masked, `"secret"`) || strings.Contains(masked, `"pass"`) {
		t.Fatalf("expected fallback redaction in %s", masked)
	}
}
