package authpublic

import (
	"crypto/sha256"
	"encoding/hex"
	"slices"
	"strings"

	"github.com/OliveTin/OliveTin/internal/config"
	log "github.com/sirupsen/logrus"
)

// User represents a person.
type AuthenticatedUser struct {
	EffectivePolicy *config.ConfigurationPolicy
	Username        string
	UsergroupLine   string
	Provider        string
	SID             string
	Acls            []string
}

// The session ID is the credential held in the session cookie, so it is never
// returned to clients. The fingerprint lets a client tell sessions apart, and
// cannot be used to authenticate.
const sessionFingerprintLength = 16

func (u *AuthenticatedUser) SessionFingerprint() string {
	if u.SID == "" {
		return ""
	}

	sum := sha256.Sum256([]byte(u.SID))

	return hex.EncodeToString(sum[:])[:sessionFingerprintLength]
}

func (u *AuthenticatedUser) IsGuest() bool {
	return u.Username == "guest" && u.Provider == "system"
}

// MustLogIn is true for unauthenticated visitors when authRequireGuestsToLogin
// is enabled. Such visitors get no permissions at all, whatever the ACLs say.
func (u *AuthenticatedUser) MustLogIn(cfg *config.Config) bool {
	return u != nil && u.IsGuest() && cfg.AuthRequireGuestsToLogin
}

func (u *AuthenticatedUser) parseUsergroupLine(sep string) []string {
	ret := []string{}

	if sep != "" {
		for v := range strings.SplitSeq(u.UsergroupLine, sep) {
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
	u.Acls = append(u.Acls, u.matchingAclNames(cfg)...)
	u.EffectivePolicy = getEffectivePolicy(cfg, u)
}

// An ACL that names the "guest" user or usergroup is meant for logged-in users
// (for example OAuth2 users with no group claim). When login is required it must
// never match an unauthenticated visitor, who is also called "guest".
func (u *AuthenticatedUser) matchingAclNames(cfg *config.Config) []string {
	if u.MustLogIn(cfg) {
		return nil
	}

	var names []string

	for _, acl := range cfg.AccessControlLists {
		if u.matchesAcl(acl, cfg.AuthHttpHeaderUserGroupSep) {
			names = append(names, acl.Name)
		}
	}

	return names
}

func (u *AuthenticatedUser) matchesAcl(acl *config.AccessControlList, usergroupSep string) bool {
	return slices.Contains(acl.MatchUsernames, u.Username) || u.MatchesUsergroupAcl(acl.MatchUsergroups, usergroupSep)
}

func getEffectivePolicy(cfg *config.Config, u *AuthenticatedUser) *config.ConfigurationPolicy {
	ret := &config.ConfigurationPolicy{
		ShowDiagnostics:   cfg.DefaultPolicy.ShowDiagnostics,
		ShowLogList:       cfg.DefaultPolicy.ShowLogList,
		ShowVersionNumber: cfg.DefaultPolicy.ShowVersionNumber,
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

	if policy.ShowVersionNumber {
		ret.ShowVersionNumber = policy.ShowVersionNumber
	}

	return ret
}
