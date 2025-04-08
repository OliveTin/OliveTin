package acl

import "testing"

func Test_hasGroupsMatch(t *testing.T) {
	tests := []struct {
		name            string
		matchUsergroups []string
		usergroup       string
		want            bool
	}{
		{
			name:            "No groups match",
			matchUsergroups: []string{"group1", "group2"},
			usergroup:       "group3",
		},
		{
			name:            "Exact match",
			matchUsergroups: []string{"group1", "group2"},
			usergroup:       "group1",
			want:            true,
		},
		{
			name:            "Multiple groups match",
			matchUsergroups: []string{"group1", "group2"},
			usergroup:       "group1 group2",
			want:            true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasGroupsMatch(tt.matchUsergroups, tt.usergroup); got != tt.want {
				t.Errorf("hasGroupsMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}
