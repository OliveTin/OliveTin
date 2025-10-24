package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindAction(t *testing.T) {
	c := DefaultConfig()

	a1 := &Action{}
	a1.Title = "a1"
	c.Actions = append(c.Actions, a1)

	a2 := &Action{
		Title: "a2",
		Arguments: []ActionArgument{
			{
				Name: "Blat",
			},
		},
	}

	c.Actions = append(c.Actions, a2)

	assert.NotNil(t, c.findAction("a1"), "Find action a1")

	assert.NotNil(t, c.findAction("a2"), "Find action a2")
	assert.NotNil(t, c.findAction("a2").FindArg("Blat"), "Find action argument")
	assert.Nil(t, c.findAction("a2").FindArg("Blatey Cake"), "Find non-existent action argument")

	assert.Nil(t, c.findAction("waffles"), "Find non-existent action")
}

func TestFindAcl(t *testing.T) {
	c := DefaultConfig()

	acl1 := &AccessControlList{
		Name: "Testing ACL",
	}

	c.AccessControlLists = append(c.AccessControlLists, acl1)

	assert.NotNil(t, c.FindAcl("Testing ACL"), "Find a ACL that should exist")
	assert.Nil(t, c.FindAcl("Chocolate Cake"), "Find a ACL that does not exist")
}

func TestSetDir(t *testing.T) {
	c := DefaultConfig()
	c.SetDir("test")

	assert.Equal(t, "test", c.GetDir(), "SetDir")
}

func TestFindUserByUsername(t *testing.T) {
	c := DefaultConfig()

	// Test with empty users list
	assert.Nil(t, c.FindUserByUsername("nonexistent"), "Find user in empty list should return nil")

	// Add test users
	user1 := &LocalUser{
		Username:  "admin",
		Usergroup: "admin",
		Password:  "adminpass",
	}
	user2 := &LocalUser{
		Username:  "guest",
		Usergroup: "guest",
		Password:  "guestpass",
	}

	c.AuthLocalUsers.Users = append(c.AuthLocalUsers.Users, user1, user2)

	// Test finding existing users
	foundUser := c.FindUserByUsername("admin")
	assert.NotNil(t, foundUser, "Find existing user 'admin'")
	assert.Equal(t, "admin", foundUser.Username, "Found user should have correct username")
	assert.Equal(t, "admin", foundUser.Usergroup, "Found user should have correct usergroup")
	assert.Equal(t, "adminpass", foundUser.Password, "Found user should have correct password")

	foundUser = c.FindUserByUsername("guest")
	assert.NotNil(t, foundUser, "Find existing user 'guest'")
	assert.Equal(t, "guest", foundUser.Username, "Found user should have correct username")
	assert.Equal(t, "guest", foundUser.Usergroup, "Found user should have correct usergroup")

	// Test finding non-existent user
	assert.Nil(t, c.FindUserByUsername("nonexistent"), "Find non-existent user should return nil")
	assert.Nil(t, c.FindUserByUsername(""), "Find empty username should return nil")
}
