package object

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
	_ "github.com/lib/pq" // db = postgres
	"github.com/xorm-io/xorm"
)

// groupTestRBACModel is a minimal RBAC model, independent of the Casdoor built-in
// enforcer bootstrap (which needs its own DB-backed model/adapter rows): it only
// needs a "g" role_definition so UserGroupEnforcer.checkModel passes.
const groupTestRBACModel = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`

// TestGetGroupUserCountReservedUserTable reproduces the bug fixed in
// GetGroupUserCount/GetPaginationGroupUsers: the field filter built a raw
// "user.<col> like ?" (or, for GetPaginationGroupUsers, "<prefixedTable>.<col> like ?")
// condition without quoting. On PostgreSQL "user" is a reserved keyword, so an
// unquoted table-qualified column reference fails with a syntax error. Needs a
// disposable Postgres instance via TEST_POSTGRES_DSN; skipped otherwise.
func TestGetGroupUserCountReservedUserTable(t *testing.T) {
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN not set, skipping Postgres-specific test")
	}

	engine, err := xorm.NewEngine("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()

	prevOrmer, prevUserEnforcer := ormer, userEnforcer
	defer func() { ormer, userEnforcer = prevOrmer, prevUserEnforcer }()

	ormer = &Ormer{Engine: engine}

	if err := engine.Sync2(new(User)); err != nil {
		t.Fatal(err)
	}
	defer engine.DropTables(new(User))

	m, err := model.NewModelFromString(groupTestRBACModel)
	if err != nil {
		t.Fatal(err)
	}
	// File adapter backed by a scratch file: GetAllUsersByGroup calls LoadPolicy on
	// every lookup, which would otherwise wipe the in-memory roles added below (the
	// adapter doesn't implement incremental AddPolicy, so those roles never reach
	// storage on their own - an explicit SavePolicy after the fixtures below is
	// what makes the later LoadPolicy read them back).
	policyPath := filepath.Join(t.TempDir(), "policy.csv")
	if err := os.WriteFile(policyPath, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	enforcer, err := casbin.NewEnforcer(m, fileadapter.NewAdapter(policyPath))
	if err != nil {
		t.Fatal(err)
	}
	userEnforcer = NewUserGroupEnforcer(&casbin.SyncedEnforcer{Enforcer: enforcer})

	const owner = "built-in"
	const groupName = "group-grpq"
	groupId := owner + "/" + groupName

	fixtures := []*User{
		{Owner: owner, Name: "alice", Phone: "5550101"},
		{Owner: owner, Name: "bob", Phone: "5550202"},
	}
	for _, u := range fixtures {
		if _, err := engine.Insert(u); err != nil {
			t.Fatal(err)
		}
		if _, err := userEnforcer.AddGroupForUser(owner+"/"+u.Name, groupId); err != nil {
			t.Fatal(err)
		}
	}
	if err := enforcer.SavePolicy(); err != nil {
		t.Fatal(err)
	}

	count, err := GetGroupUserCount(groupId, "phone", "0101")
	if err != nil {
		t.Fatalf("GetGroupUserCount with a field filter on the reserved \"user\" table: %v", err)
	}
	if count != 1 {
		t.Fatalf("GetGroupUserCount(phone=0101) = %d, want 1", count)
	}

	filtered, err := GetPaginationGroupUsers(groupId, -1, -1, "phone", "0101", "", "")
	if err != nil {
		t.Fatalf("GetPaginationGroupUsers with a field filter on the reserved \"user\" table: %v", err)
	}
	if len(filtered) != 1 || filtered[0].Name != "alice" {
		t.Fatalf("GetPaginationGroupUsers(phone=0101) = %v, want [alice]", filtered)
	}
}
