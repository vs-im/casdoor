package object

import (
	"reflect"
	"testing"
)

func TestGetTokenAudienceSharedApp(t *testing.T) {
	application := &Application{
		Owner:        "admin",
		Name:         "app-hasura",
		Organization: "built-in",
		ClientId:     "client123",
		IsShared:     true,
	}

	t.Run("native organization keeps plain client_id", func(t *testing.T) {
		user := &User{Owner: "built-in", Name: "admin"}
		aud := getTokenAudience(application, user, "")
		expected := []string{"client123"}
		if !reflect.DeepEqual([]string(aud), expected) {
			t.Fatalf("expected aud %v, got %v", expected, aud)
		}
	})

	t.Run("foreign organization gets plain and org-suffixed client_id", func(t *testing.T) {
		user := &User{Owner: "acme", Name: "alice"}
		aud := getTokenAudience(application, user, "")
		expected := []string{"client123", "client123-org-acme"}
		if !reflect.DeepEqual([]string(aud), expected) {
			t.Fatalf("expected aud %v, got %v", expected, aud)
		}
	})
}

func TestGetTokenAudienceNonSharedApp(t *testing.T) {
	application := &Application{
		Owner:        "admin",
		Name:         "app-plain",
		Organization: "acme",
		ClientId:     "client456",
		IsShared:     false,
	}

	user := &User{Owner: "acme", Name: "alice"}
	aud := getTokenAudience(application, user, "")
	expected := []string{"client456"}
	if !reflect.DeepEqual([]string(aud), expected) {
		t.Fatalf("expected aud %v, got %v", expected, aud)
	}
}

func TestGetTokenAudienceResourceOverride(t *testing.T) {
	application := &Application{
		Owner:        "admin",
		Name:         "app-hasura",
		Organization: "built-in",
		ClientId:     "client123",
		IsShared:     true,
	}

	user := &User{Owner: "acme", Name: "alice"}
	aud := getTokenAudience(application, user, "https://api.example.com")
	expected := []string{"https://api.example.com"}
	if !reflect.DeepEqual([]string(aud), expected) {
		t.Fatalf("expected aud %v, got %v", expected, aud)
	}
}
