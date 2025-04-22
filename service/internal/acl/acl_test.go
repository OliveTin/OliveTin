package acl

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
			user := &AuthenticatedUser{
				Username:      "testuser",
				UsergroupLine: tt.usergroupLine,
			}

			if matches := user.matchesUsergroupAcl(tt.aclMatchUsergroups, tt.sep); matches != tt.matches {
				t.Errorf("AuthenticatedUser.matchesUsergroupAcl() = %v, want %v for usergroups %v", matches, tt.matches, tt.aclMatchUsergroups)
			}
		})
	}
}

func Test_parseUsergroupLine(t *testing.T) {
	tests := []struct {
		name           string
		usergroupLine  string
		expectedGroups []string
		sep            string
	}{
		{
			name:           "Default separator (space)",
			usergroupLine:  "group1 group2",
			expectedGroups: []string{"group1", "group2"},
		},
		{
			name:           "Comma-separated groups",
			usergroupLine:  "group1 , group2",
			expectedGroups: []string{"group1", "group2"},
			sep:            ",",
		},
		{
			name:           "Multiple spaces",
			usergroupLine:  "group1 , group2      , group3",
			expectedGroups: []string{"group1", "group2", "group3"},
			sep:            ",",
		},
		{
			name:           "Empty usergroup line",
			usergroupLine:  "",
			expectedGroups: []string{},
		},
		{
			name:           "Empty group names",
			usergroupLine:  "|group1| | group3|",
			expectedGroups: []string{"group1", "group3"},
			sep:            "|",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &AuthenticatedUser{
				Username:      "testuser",
				UsergroupLine: tt.usergroupLine,
			}

			assert.Equal(t, tt.expectedGroups, user.parseUsergroupLine(tt.sep))
		})
	}
}
