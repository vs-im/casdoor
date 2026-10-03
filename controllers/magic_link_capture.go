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
	"encoding/json"
	"strings"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// The dev side of the magic link capture (object/magic_link_capture.go): a test server
// holding the application's client credentials reads the open link that was not mailed.

type CapturedMagicLinkRequestForm struct {
	Email        string `json:"email"`
	Organization string `json:"organization"`
	Application  string `json:"application"`
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

// GetCapturedMagicLink ...
// @Title GetCapturedMagicLink
// @Tag Magic Link API
// @Description dev only (CASDOOR_DEV_MAGIC_LINK_CAPTURE=true, 404 otherwise): the last open magic link captured for a ".test" address, needs the client credentials of the application
// @Param form body controllers.CapturedMagicLinkRequestForm true "Captured magic link request"
// @Success 200 {object} controllers.Response The Response object
// @router /get-captured-magic-link [post]
func (c *ApiController) GetCapturedMagicLink() {
	if !object.IsDevMagicLinkCaptureEnabled() {
		c.Ctx.Output.SetStatus(404)
		c.Ctx.WriteString("Not Found")
		return
	}

	var requestForm CapturedMagicLinkRequestForm
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &requestForm)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	requestForm.Email = strings.TrimSpace(strings.ToLower(requestForm.Email))
	requestForm.Organization = strings.TrimSpace(requestForm.Organization)
	requestForm.Application = strings.TrimSpace(requestForm.Application)
	requestForm.ClientID = strings.TrimSpace(requestForm.ClientID)
	requestForm.ClientSecret = strings.TrimSpace(requestForm.ClientSecret)
	if requestForm.Email == "" || requestForm.Organization == "" {
		c.ResponseError(c.T("general:Missing parameter"))
		return
	}

	application, err := c.getMagicLinkApplication(&MagicLinkRequestForm{
		Organization: requestForm.Organization,
		Application:  requestForm.Application,
	})
	if err != nil || application == nil || application.Organization != requestForm.Organization {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}
	ok, err := validateMagicLinkClientCredentials(application, requestForm.ClientID, requestForm.ClientSecret)
	if err != nil || !ok {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}

	link, found := object.GetCapturedMagicLink(util.GetId(application.Organization, application.Name), requestForm.Email)
	if !found {
		c.ResponseError("no captured magic link")
		return
	}
	c.ResponseOk(map[string]string{"link": link})
}
