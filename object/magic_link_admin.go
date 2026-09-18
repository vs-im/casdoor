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
	"time"

	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

// The admin side of the magic links: the list page, revoking and deleting a link.

// fillBuiltInMagicLinkStatus gives a link of the built-in sign-in page, which has neither a
// status nor an expiry of its own, the ones the list page shows.
func fillBuiltInMagicLinkStatus(links []*MagicLink) {
	now := time.Now().Unix()
	for _, link := range links {
		if link.Status != "" {
			continue
		}

		expireAt := time.Unix(link.Time+getMagicLinkTimeout()*60, 0)
		link.ExpireAt = expireAt.Unix()
		link.ExpiryTime = util.Time2String(expireAt)
		if link.IsUsed {
			link.Status = MagicLinkStatusUsed
		} else if now > link.ExpireAt {
			link.Status = MagicLinkStatusExpired
		} else {
			link.Status = MagicLinkStatusSent
		}
	}
}

func GetMagicLinkCount(owner, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&MagicLink{Owner: owner})
}

func GetMagicLinks(owner string) ([]*MagicLink, error) {
	links := []*MagicLink{}
	err := ormer.Engine.Desc("created_time").Find(&links, &MagicLink{Owner: owner})
	fillBuiltInMagicLinkStatus(links)
	return links, err
}

func GetPaginationMagicLinks(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*MagicLink, error) {
	links := []*MagicLink{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&links, &MagicLink{Owner: owner})
	fillBuiltInMagicLinkStatus(links)
	return links, err
}

func GetMagicLink(id string) (*MagicLink, error) {
	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)
	link := MagicLink{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&link)
	if err != nil {
		return nil, err
	}
	if !existed {
		return nil, nil
	}
	return &link, nil
}

// RevokeMagicLink takes a link that was not used yet out of circulation. "is_used" is
// what ConsumeMagicLink() claims on, so a revoked link is dead for the built-in sign-in
// page and for the API alike, and revoking cannot race a sign-in that already won it.
func RevokeMagicLink(id string) (bool, error) {
	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)
	link := &MagicLink{IsUsed: true}
	link.Status = MagicLinkStatusRevoked
	affected, err := ormer.Engine.ID(core.PK{owner, name}).And("is_used = ?", false).Cols("is_used", "status").Update(link)
	return affected != 0, err
}

func DeleteMagicLink(id string) (bool, error) {
	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)
	affected, err := ormer.Engine.ID(core.PK{owner, name}).Delete(&MagicLink{})
	return affected != 0, err
}
