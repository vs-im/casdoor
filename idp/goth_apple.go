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

package idp

import (
	"fmt"
	"reflect"

	"github.com/markbates/goth"
	"golang.org/x/oauth2"
)

// forceAppleBodyCredentials makes the Sign in with Apple code exchange a single POST with client_id and
// client_secret in the form body only.
//
// goth's apple provider builds its oauth2.Config without Endpoint.AuthStyle, so golang.org/x/oauth2 auto-detects:
// it first POSTs with an Authorization: Basic header and, on any error, POSTs the same authorization code again
// with the credentials in the body. Apple accepts a code once, so the second POST always fails with invalid_grant
// and hides the answer to the first one. The config is unexported in goth and Session.Authorize does not take a
// client, so the style is set on the config in place.
func forceAppleBodyCredentials(p goth.Provider) error {
	cfg, err := appleOAuth2Config(p)
	if err != nil {
		return err
	}
	cfg.Endpoint.AuthStyle = oauth2.AuthStyleInParams
	return nil
}

// appleOAuth2Config returns the oauth2.Config held by goth's apple provider.
func appleOAuth2Config(p goth.Provider) (*oauth2.Config, error) {
	v := reflect.ValueOf(p)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return nil, fmt.Errorf("apple provider: unexpected type %T", p)
	}
	field := v.Elem().FieldByName("config")
	if !field.IsValid() || field.Type() != reflect.TypeOf((*oauth2.Config)(nil)) || field.IsNil() {
		return nil, fmt.Errorf("apple provider: oauth2 config not found, goth layout changed")
	}
	return (*oauth2.Config)(field.UnsafePointer()), nil
}
