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
	"fmt"

	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

// The permission a link of the magic link API signs in under, and its snapshot in the row.

func ApplyPermissionSnapshotToMagicLink(link *MagicLink, permission *Permission) {
	if permission == nil {
		return
	}
	link.Permission = permission.GetId()
	link.SubUsers = append([]string{}, permission.Users...)
	link.SubGroups = append([]string{}, permission.Groups...)
	link.SubRoles = append([]string{}, permission.Roles...)
	link.SubDomains = append([]string{}, permission.Domains...)
	link.Resources = append([]string{}, permission.Resources...)
	link.Actions = append([]string{}, permission.Actions...)
}

func UpdateMagicLinkPermissionSnapshot(link *MagicLink) error {
	_, err := ormer.Engine.ID(core.PK{link.Owner, link.Name}).Cols("permission", "sub_users", "sub_groups", "sub_roles", "sub_domains", "resources", "actions").Update(link)
	return err
}

func ResolveMagicLinkPermission(application *Application, user *User, requestedPermission string) (*Permission, error) {
	owner := application.Organization
	if application.IsShared && user != nil {
		owner = user.Owner
	}
	if requestedPermission != "" {
		permissionOwner, permissionName := util.GetOwnerAndNameFromIdNoCheck(requestedPermission)
		if permissionName == "" {
			permissionOwner = owner
			permissionName = requestedPermission
		}
		if permissionOwner != owner {
			return nil, fmt.Errorf("permission company mismatch")
		}
		permission, err := getPermission(permissionOwner, permissionName)
		if err != nil {
			return nil, err
		}
		if permission == nil {
			return nil, fmt.Errorf("permission does not exist")
		}
		return validateMagicLinkPermission(application, user, permission)
	}

	permissions, err := GetPermissions(owner)
	if err != nil {
		return nil, err
	}
	for _, permission := range permissions {
		if _, err = validateMagicLinkPermission(application, user, permission); err == nil {
			return permission, nil
		}
	}
	if user != nil {
		ok, err := CheckLoginPermission(user.GetId(), application)
		if err != nil {
			return nil, err
		}
		if ok {
			return nil, nil
		}
	}
	return nil, fmt.Errorf("magic link login permission denied")
}

func validateMagicLinkPermission(application *Application, user *User, permission *Permission) (*Permission, error) {
	if permission == nil || !permission.IsEnabled || permission.State != "Approved" || permission.ResourceType != "Application" || !permission.isResourceHit(application.Name) || permission.Effect != "Allow" {
		return nil, fmt.Errorf("permission does not allow this application")
	}
	if user == nil {
		return permission, nil
	}
	userID := user.GetId()
	userHit := permission.isUserHit(userID)
	groupHit := permission.isGroupHit(userID)
	roleHit := permission.isRoleHit(userID)
	if !userHit && !groupHit && !roleHit {
		return nil, fmt.Errorf("permission does not allow this user")
	}
	enforcer, err := getPermissionEnforcer(permission)
	if err != nil {
		return nil, err
	}
	ok, err := enforcer.Enforce(userID, application.Name, "Read")
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("permission does not allow this action")
	}
	return permission, nil
}
