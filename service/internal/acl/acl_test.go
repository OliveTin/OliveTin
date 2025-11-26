package acl

import (
	"testing"

	authpublic "github.com/OliveTin/OliveTin/internal/auth/authpublic"
)

func Test_hasGroupsMatch(t *testing.T) {
	tests := []struct {
		name               string
		aclMatchUsergroups []string
		usergroupLine      string
		matches            bool
		sep                string
	}{
		{
			name:               "No groups match",
			aclMatchUsergroups: []string{"group1", "group2"},
			usergroupLine:      "group3",
			matches:            false,
		},
		{
			name:               "Exact match",
			aclMatchUsergroups: []string{"group1", "group2"},
			usergroupLine:      "group1",
			matches:            true,
		},
		{
			name:               "Multiple groups match",
			aclMatchUsergroups: []string{"group1", "group2"},
			usergroupLine:      "group1 group2",
			matches:            true,
		},
		{
			name:               "Comma-separated groups match",
			aclMatchUsergroups: []string{"group1", "group2", "group3"},
			usergroupLine:      "group1, group2",
			matches:            true,
			sep:                ",",
		},
		{
			name:               "Comma-separated groups with default separator does not match",
			aclMatchUsergroups: []string{"group1"},
			usergroupLine:      "group1, group2",
			matches:            false,
			sep:                "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &authpublic.AuthenticatedUser{
				Username:      "testuser",
				UsergroupLine: tt.usergroupLine,
			}

			if matches := user.MatchesUsergroupAcl(tt.aclMatchUsergroups, tt.sep); matches != tt.matches {
				t.Errorf("AuthenticatedUser.MatchesUsergroupAcl() = %v, want %v for usergroups %v", matches, tt.matches, tt.aclMatchUsergroups)
			}
		})
	}
}
