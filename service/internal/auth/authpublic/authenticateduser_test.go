package authpublic

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
