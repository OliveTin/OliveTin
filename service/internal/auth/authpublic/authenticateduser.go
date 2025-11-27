package authpublic

import (
	"slices"
	"strings"

	"github.com/OliveTin/OliveTin/internal/config"
	log "github.com/sirupsen/logrus"
)

// User represents a person.
type AuthenticatedUser struct {
	Username      string
	UsergroupLine string

	Provider string
	SID      string

	Acls []string

	EffectivePolicy *config.ConfigurationPolicy
}

func (u *AuthenticatedUser) IsGuest() bool {
	return u.Username == "guest" && u.Provider == "system"
}

func (u *AuthenticatedUser) parseUsergroupLine(sep string) []string {
	ret := []string{}

	if sep != "" {
		for _, v := range strings.Split(u.UsergroupLine, sep) {
			trimmed := strings.TrimSpace(v)

			if trimmed != "" {
				ret = append(ret, trimmed)
			}
		}
	} else {
		ret = strings.Fields(u.UsergroupLine)
	}

	log.Debugf("parseUsergroupLine: %v, %v, sep:%v", u.UsergroupLine, ret, sep)

	return ret
}

func (u *AuthenticatedUser) MatchesUsergroupAcl(matchUsergroups []string, sep string) bool {
	groupList := u.parseUsergroupLine(sep)

	for _, group := range groupList {
		if slices.Contains(matchUsergroups, group) {
			log.Debugf("Usergroup %v found in %+v (len: %v)", group, groupList, len(groupList))
			return true
		}
	}

	return false
}

func (u *AuthenticatedUser) BuildUserAcls(cfg *config.Config) {
	for _, acl := range cfg.AccessControlLists {
		if slices.Contains(acl.MatchUsernames, u.Username) {
			u.Acls = append(u.Acls, acl.Name)
			continue
		}

		if u.MatchesUsergroupAcl(acl.MatchUsergroups, cfg.AuthHttpHeaderUserGroupSep) {
			u.Acls = append(u.Acls, acl.Name)
			continue
		}
	}

	u.EffectivePolicy = getEffectivePolicy(cfg, u)
}

func getEffectivePolicy(cfg *config.Config, u *AuthenticatedUser) *config.ConfigurationPolicy {
	ret := &config.ConfigurationPolicy{
		ShowDiagnostics: cfg.DefaultPolicy.ShowDiagnostics,
		ShowLogList:     cfg.DefaultPolicy.ShowLogList,
	}

	for _, acl := range cfg.AccessControlLists {
		if slices.Contains(u.Acls, acl.Name) {
			ret = buildConfigurationPolicy(ret, acl.Policy)
		}
	}

	return ret
}

func buildConfigurationPolicy(ret *config.ConfigurationPolicy, policy config.ConfigurationPolicy) *config.ConfigurationPolicy {
	if policy.ShowDiagnostics {
		ret.ShowDiagnostics = policy.ShowDiagnostics
	}

	if policy.ShowLogList {
		ret.ShowLogList = policy.ShowLogList
	}

	return ret
}
