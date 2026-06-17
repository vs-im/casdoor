package controllers

import (
	"fmt"
	"strings"

	"github.com/casdoor/casdoor/object"
)

func resolveSignupUsername(application *object.Application, organization *object.Organization, requestedUsername string, email string, generatedID string) string {
	username := strings.TrimSpace(requestedUsername)
	if username != "" && application.IsSignupItemVisible("Username") {
		return username
	}
	if organization != nil && organization.UseEmailAsUsername && email != "" {
		return strings.ToLower(strings.TrimSpace(email))
	}
	return generatedID
}

func resolveSignupDisplayName(email string, username string) string {
	email = strings.TrimSpace(strings.ToLower(email))
	if email != "" {
		return email
	}
	return strings.TrimSpace(username)
}

func (c *ApiController) createSignupUser(user *object.User, allowExistingByEmail bool) (*object.User, bool, error) {
	affected, err := object.AddUser(user, c.GetAcceptLanguage())
	if err == nil && affected {
		err = object.AddUserToOriginalDatabase(user)
		if err != nil {
			return nil, false, err
		}
		return user, true, nil
	}
	if allowExistingByEmail {
		existedUser, getErr := object.GetUserByEmail(user.Owner, strings.ToLower(strings.TrimSpace(user.Email)))
		if getErr != nil {
			if err != nil {
				return nil, false, err
			}
			return nil, false, getErr
		}
		if existedUser != nil {
			return existedUser, false, nil
		}
	}
	if err != nil {
		return nil, false, err
	}
	return nil, false, fmt.Errorf("failed to add user")
}
