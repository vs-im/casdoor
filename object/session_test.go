package object

import "testing"

// TestNormalizeSessionFilterField — фильтр get-sessions по field=user должен уходить
// в колонку name (у таблицы session нет колонки user: пользователь сессии — pk name).
func TestNormalizeSessionFilterField(t *testing.T) {
	if got := normalizeSessionFilterField("user"); got != "name" {
		t.Fatalf("normalizeSessionFilterField(user) = %q, want name", got)
	}
	for _, field := range []string{"name", "application", "createdTime", ""} {
		if got := normalizeSessionFilterField(field); got != field {
			t.Fatalf("normalizeSessionFilterField(%q) = %q, want unchanged", field, got)
		}
	}
}
