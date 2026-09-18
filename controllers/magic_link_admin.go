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

package controllers

import (
	"github.com/beego/beego/v2/core/utils/pagination"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// The admin side of the magic links: the list page, revoking and deleting a link.

// GetMagicLinks ...
// @Title GetMagicLinks
// @Tag Magic Link API
// @Description get Magic Link records for admin UI
// @Param owner query string false "The owner of Magic Link records"
// @Param p query string false "The page number"
// @Param pageSize query string false "The page size"
// @Param field query string false "The search field"
// @Param value query string false "The search value"
// @Param sortField query string false "The sort field"
// @Param sortOrder query string false "The sort order"
// @Success 200 {array} object.MagicLink The Response object
// @router /get-magic-links [get]
func (c *ApiController) GetMagicLinks() {
	organization, ok := c.RequireAdmin()
	if !ok {
		return
	}

	limit := c.Ctx.Input.Query("pageSize")
	page := c.Ctx.Input.Query("p")
	field := c.Ctx.Input.Query("field")
	value := c.Ctx.Input.Query("value")
	status := c.Ctx.Input.Query("status")
	user := c.Ctx.Input.Query("user")
	email := c.Ctx.Input.Query("email")
	application := c.Ctx.Input.Query("application")
	queryOrganization := c.Ctx.Input.Query("organization")
	group := c.Ctx.Input.Query("group")
	permission := c.Ctx.Input.Query("permission")
	sortField := c.Ctx.Input.Query("sortField")
	sortOrder := c.Ctx.Input.Query("sortOrder")
	owner := c.Ctx.Input.Query("owner")
	if owner == "" && queryOrganization != "" {
		owner = queryOrganization
	}
	if c.IsGlobalAdmin() && owner != "" {
		organization = owner
	}
	if field == "" || value == "" {
		filterPairs := []struct {
			field string
			value string
		}{
			{field: "status", value: status},
			{field: "requester", value: user},
			{field: "email", value: email},
			{field: "application", value: application},
			{field: "owner", value: queryOrganization},
			{field: "subGroups", value: group},
			{field: "permission", value: permission},
		}
		for _, pair := range filterPairs {
			if pair.value != "" {
				field = pair.field
				value = pair.value
				break
			}
		}
	}

	if limit == "" || page == "" {
		links, err := object.GetMagicLinks(organization)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		c.ResponseOk(links)
	} else {
		limit := util.ParseInt(limit)
		count, err := object.GetMagicLinkCount(organization, field, value)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		paginator := pagination.NewPaginator(c.Ctx.Request, limit, count)
		links, err := object.GetPaginationMagicLinks(organization, paginator.Offset(), limit, field, value, sortField, sortOrder)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		c.ResponseOk(links, paginator.Nums())
	}
}

// RevokeMagicLink ...
// @Title RevokeMagicLink
// @Tag Magic Link API
// @Description revoke an unused Magic Link
// @Param id query string true "The id (owner/name) of the Magic Link"
// @Success 200 {object} controllers.Response The Response object
// @router /revoke-magic-link [post]
func (c *ApiController) RevokeMagicLink() {
	organization, ok := c.RequireAdmin()
	if !ok {
		return
	}

	id := c.Ctx.Input.Query("id")
	if id == "" {
		c.ResponseError(c.T("general:Missing parameter"))
		return
	}
	owner, _ := util.GetOwnerAndNameFromIdNoCheck(id)
	if !c.IsGlobalAdmin() && owner != organization {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}
	c.Data["json"] = wrapActionResponse(object.RevokeMagicLink(id))
	c.ServeJSON()
}

// DeleteMagicLink ...
// @Title DeleteMagicLink
// @Tag Magic Link API
// @Description delete a Magic Link record
// @Param id query string true "The id (owner/name) of the Magic Link"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-magic-link [post]
func (c *ApiController) DeleteMagicLink() {
	organization, ok := c.RequireAdmin()
	if !ok {
		return
	}

	id := c.Ctx.Input.Query("id")
	if id == "" {
		c.ResponseError(c.T("general:Missing parameter"))
		return
	}
	owner, _ := util.GetOwnerAndNameFromIdNoCheck(id)
	if !c.IsGlobalAdmin() && owner != organization {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}
	c.Data["json"] = wrapActionResponse(object.DeleteMagicLink(id))
	c.ServeJSON()
}
