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
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/xorm-io/core"
)

// The "magic_link" table of the fork is older than the built-in one. Sync2() cannot add
// the built-in NOT NULL columns to a table that already has rows, so they are added here
// with a default first, and the old rows are then moved to the built-in model: the
// application by its name, the issue time, "is_used" and the session hash of a link that
// may be opened anywhere. Every step only touches what is still in the old shape.

var magicLinkBuiltInColumns = []struct {
	name    string
	sqlType map[string]string
}{
	{"session_hash", map[string]string{"": "VARCHAR(100) NOT NULL DEFAULT ''"}},
	{"time", map[string]string{"": "BIGINT NOT NULL DEFAULT 0"}},
	{"is_used", map[string]string{"": "BOOLEAN NOT NULL DEFAULT false", "mysql": "TINYINT(1) NOT NULL DEFAULT 0", "mssql": "BIT NOT NULL DEFAULT 0", "sqlite3": "INTEGER NOT NULL DEFAULT 0"}},
}

func (a *Ormer) syncMagicLink() error {
	err := a.prepareLegacyMagicLinkTable()
	if err != nil {
		return err
	}

	err = a.Engine.Sync2(new(MagicLink))
	if err != nil {
		return err
	}

	return a.migrateLegacyMagicLinks()
}

func (a *Ormer) prepareLegacyMagicLinkTable() error {
	tableName := a.Engine.TableName(new(MagicLink))
	existed, err := a.Engine.IsTableExist(tableName)
	if err != nil || !existed {
		return err
	}

	dbType := string(a.Engine.Dialect().URI().DBType)
	for _, column := range magicLinkBuiltInColumns {
		existed, err = a.Engine.Dialect().IsColumnExist(a.Engine.DB(), context.Background(), tableName, column.name)
		if err != nil {
			return err
		}
		if existed {
			continue
		}

		sqlType, ok := column.sqlType[dbType]
		if !ok {
			sqlType = column.sqlType[""]
		}
		// another instance starting against the same database may add the column in between
		addColumn := "ADD"
		if dbType == "postgres" {
			addColumn = "ADD COLUMN IF NOT EXISTS"
		}
		_, err = a.Engine.Exec(fmt.Sprintf("ALTER TABLE %s %s %s %s", a.Engine.Quote(tableName), addColumn, a.Engine.Quote(column.name), sqlType))
		if err != nil {
			added, checkErr := a.Engine.Dialect().IsColumnExist(a.Engine.DB(), context.Background(), tableName, column.name)
			if checkErr != nil || !added {
				return err
			}
		}
	}
	return nil
}

func getLegacyMagicLinkTime(link *MagicLink, now int64) int64 {
	issued, err := time.Parse(time.RFC3339, link.CreatedTime)
	if err == nil && issued.Unix() > 0 && issued.Unix() <= now {
		return issued.Unix()
	}
	if link.ExpireAt > 0 && link.ExpireAt <= now {
		return link.ExpireAt
	}
	return now
}

func convertLegacyMagicLink(link *MagicLink, now int64) {
	if strings.HasPrefix(link.Application, "admin/") {
		link.Application = strings.TrimPrefix(link.Application, "admin/")
	}
	link.Time = getLegacyMagicLinkTime(link, now)
	link.IsUsed = link.TokenHash == "" || !isMagicLinkVerifiableStatus(link.Status) || link.Status == ""
	link.Binding = MagicLinkBindingNone
	link.SessionHash = ""
	if link.TokenHash != "" {
		link.SessionHash = getUnboundMagicLinkSessionHash(link.TokenHash)
	}
}

func (a *Ormer) migrateLegacyMagicLinks() error {
	now := time.Now().Unix()
	for {
		links := []*MagicLink{}
		// a row of the old model has no issue time, but always has an expiry
		err := a.Engine.Where("time = ?", 0).And("expire_at > ?", 0).Limit(500).Find(&links)
		if err != nil {
			return err
		}
		if len(links) == 0 {
			return nil
		}

		for _, link := range links {
			convertLegacyMagicLink(link, now)
			_, err = a.Engine.ID(core.PK{link.Owner, link.Name}).Cols("application", "time", "is_used", "binding", "session_hash").Update(link)
			if err != nil {
				return err
			}
		}
	}
}
