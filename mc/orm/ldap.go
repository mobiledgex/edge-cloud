package orm

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/util"
	"github.com/nmcclain/ldap"
)

// LDAP interface to MC user database

const (
	OUusers = "users"
	OUorgs  = "orgs"
)

type ldapHandler struct {
}

func (s *ldapHandler) Bind(bindDN, bindSimplePw string, conn net.Conn) (ldap.LDAPResultCode, error) {
	log.DebugLog(log.DebugLevelApi, "LDAP bind",
		"bindDN", bindDN)
	dn, err := parseDN(bindDN)
	if err != nil {
		return ldap.LDAPResultInvalidDNSyntax, nil
	}
	if dn.ou == OUusers && dn.cn == "gitlab" && bindSimplePw == "gitlab" {
		return ldap.LDAPResultSuccess, nil
	}
	if dn.ou == OUusers {
		lookup := ormapi.User{Name: dn.cn}
		user := ormapi.User{}
		log.DebugLog(log.DebugLevelApi, "LDAP bind", "lookup", lookup)

		err := db.Where(&lookup).First(&user).Error
		if err != nil {
			time.Sleep(BadAuthDelay)
			return ldap.LDAPResultInvalidCredentials, err
		}
		log.DebugLog(log.DebugLevelApi, "LDAP bind pw check", "user", user)
		matches, err := PasswordMatches(bindSimplePw, user.Passhash, user.Salt, user.Iter)
		if err != nil || !matches {
			time.Sleep(BadAuthDelay)
			return ldap.LDAPResultInvalidCredentials, err
		}
		log.DebugLog(log.DebugLevelApi, "LDAP bind success", "user", user)
		return ldap.LDAPResultSuccess, nil
	}
	return ldap.LDAPResultInvalidCredentials, nil
}

func (s *ldapHandler) Search(boundDN string, searchReq ldap.SearchRequest, conn net.Conn) (ldap.ServerSearchResult, error) {
	log.DebugLog(log.DebugLevelApi, "LDAP search",
		"boundDN", boundDN,
		"req", searchReq)
	res := ldap.ServerSearchResult{}

	pkt, err := ldap.CompileFilter(searchReq.Filter)
	if err != nil {
		return res, err
	}

	// only support basic searches
	if searchReq.BaseDN == "" {
		if pkt.Tag == ldap.FilterEqualityMatch && len(pkt.Children) == 2 {
			key, kok := pkt.Children[0].Value.(string)
			val, vok := pkt.Children[1].Value.(string)
			if kok && vok && key == "sAMAccountName" {
				ldapLookupUsername(val, &res)
			}
		}
	} else {
		dn, err := parseDN(searchReq.BaseDN)
		if err != nil {
			return res, fmt.Errorf("Invalid BaseDN, %s", err.Error())
		}
		if dn.ou == OUusers {
			ldapLookupUsername(dn.cn, &res)
		} else if dn.ou == OUorgs {
			// not yet
		} else {
			return res, fmt.Errorf("Invalid OU %s", dn.ou)
		}
	}
	res.ResultCode = ldap.LDAPResultSuccess
	log.DebugLog(log.DebugLevelApi, "LDAP search result", "res", res)
	return res, nil
}

func ldapLookupUsername(username string, result *ldap.ServerSearchResult) {
	user := ormapi.User{Name: username}
	err := db.Where(&user).First(&user).Error
	if err != nil {
		return
	}
	dn := ldapdn{
		cn: username,
		ou: OUusers,
	}
	entry := ldap.Entry{
		DN: dn.String(),
		Attributes: []*ldap.EntryAttribute{
			&ldap.EntryAttribute{
				Name:   "cn",
				Values: []string{username},
			},
			&ldap.EntryAttribute{
				Name:   "sAMAccountName",
				Values: []string{username},
			},
			&ldap.EntryAttribute{
				Name:   "email",
				Values: []string{user.Email},
			},
			&ldap.EntryAttribute{
				Name:   "mail",
				Values: []string{user.Email},
			},
			&ldap.EntryAttribute{
				Name:   "userPrincipalName",
				Values: []string{username + "@" + OUusers},
			},
			&ldap.EntryAttribute{
				Name:   "objectClass",
				Values: []string{"posixAccount"},
			},
		},
	}
	roles, err := ShowUserRoleObj(username)
	if err == nil {
		orgs := []string{}
		for _, role := range roles {
			// for now any role has full access
			dn := ldapdn{
				cn: role.Org,
				ou: OUorgs,
			}
			orgs = append(orgs, dn.String())
		}
		if len(orgs) > 0 {
			attr := ldap.EntryAttribute{
				Name:   "memberof",
				Values: orgs,
			}
			entry.Attributes = append(entry.Attributes, &attr)
		}
	}

	result.Entries = append(result.Entries, &entry)
}

// Note special char handling is accomplished by disallowing
// special chars for User or Organization names.
type ldapdn struct {
	cn string // common name (unique identifier)
	ou string // organization unit (users, orgs)
}

func parseDN(str string) (ldapdn, error) {
	dn := ldapdn{}
	strs := strings.Split(str, ",")
	for _, subdn := range strs {
		subdn = util.UnescapeLDAPName(subdn)
		kv := strings.Split(subdn, "=")
		if len(kv) != 2 {
			return dn, fmt.Errorf("LDAP DN Key-value parse error for %s", str)
		}
		switch kv[0] {
		case "cn":
			dn.cn = kv[1]
		case "ou":
			dn.ou = kv[1]
		default:
			return dn, fmt.Errorf("LDAP DN invalid component %s", kv[0])
		}
	}
	return dn, nil
}

func (s *ldapdn) String() string {
	strs := []string{}
	if s.cn != "" {
		strs = append(strs, "cn="+util.EscapeLDAPName(s.cn))
	}
	if s.ou != "" {
		strs = append(strs, "ou="+util.EscapeLDAPName(s.ou))
	}
	if len(strs) == 0 {
		return ""
	}
	return strings.Join(strs, ",")
}
