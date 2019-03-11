package orm

import (
	"time"

	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/util"
	gitlab "github.com/xanzy/go-gitlab"
)

// Gitlab's groups and group members are a duplicate of the Organizations
// and Org Roles in MC. Because it's a duplicate, it's possible to get
// out of sync (either due to failed operations, or MC or gitlab DB reset
// or restored from backup, etc). GitlabSync takes care of re-syncing.
// Syncs are triggered either by a failure, or by an API call.

// Sync Interval attempts to re-sync if there was a failure
var GitlabSyncInterval = 5 * time.Minute

type GitlabSync struct {
	run       chan bool
	needsSync bool
}

func gitlabNewSync() *GitlabSync {
	sync := GitlabSync{}
	sync.run = make(chan bool, 1)
	return &sync
}

func (s *GitlabSync) Start() {
	go func() {
		for {
			time.Sleep(GitlabSyncInterval)
			if s.needsSync {
				s.wakeup()
			}
		}
	}()
	s.NeedsSync()
	s.wakeup()
	go s.runThread()
}

func (s *GitlabSync) runThread() {
	var err error
	for {
		if err != nil {
			err = nil
		}
		select {
		case <-s.run:
		}
		log.DebugLog(log.DebugLevelApi, "Gitlab Sync running")
		s.needsSync = false
		s.syncUsers()
		s.syncGroups()
		s.syncGroupMembers()
	}
}

func (s *GitlabSync) syncUsers() {
	// get Gitlab users
	gusers, _, err := gitlabClient.Users.ListUsers(&gitlab.ListUsersOptions{})
	if err != nil {
		s.syncErr(err)
		return
	}
	gusersT := make(map[string]*gitlab.User)
	for ii, _ := range gusers {
		gusersT[gusers[ii].Name] = gusers[ii]
	}
	// get MC users
	mcusers := []ormapi.User{}
	err = db.Find(&mcusers).Error
	if err != nil {
		s.syncErr(err)
		return
	}
	mcusersT := make(map[string]*ormapi.User)
	for ii, _ := range mcusers {
		mcusersT[mcusers[ii].Name] = &mcusers[ii]
	}

	for name, user := range mcusersT {
		if _, found := gusersT[name]; found {
			// in sync
			delete(gusersT, name)
		} else {
			// missing from gitlab, so create
			log.DebugLog(log.DebugLevelApi,
				"Gitlab Sync create missing LDAP user",
				"user", name)
			gitlabCreateLDAPUser(user)
		}
	}
	for _, guser := range gusersT {
		// delete extra LDAP users - first confirm it's an LDAP user
		if guser.Identities == nil {
			continue
		}
		ldapuser := false
		for _, id := range guser.Identities {
			if id.Provider == LDAPProvider {
				ldapuser = true
				break
			}
		}
		if !ldapuser {
			continue
		}
		log.DebugLog(log.DebugLevelApi,
			"Gitlab Sync delete extra LDAP user",
			"name", guser.Name)
		_, err = gitlabClient.Users.DeleteUser(guser.ID)
		if err != nil {
			s.syncErr(err)
		}
	}
}

func (s *GitlabSync) syncGroups() {
	// get Gitlab groups
	groups, _, err := gitlabClient.Groups.ListGroups(&gitlab.ListGroupsOptions{})
	if err != nil {
		s.syncErr(err)
		return
	}
	groupsT := make(map[string]*gitlab.Group)
	for ii, _ := range groups {
		groupsT[groups[ii].Name] = groups[ii]
	}
	// get MC orgs
	orgs := []ormapi.Organization{}
	err = db.Find(&orgs).Error
	if err != nil {
		s.syncErr(err)
		return
	}
	orgsT := make(map[string]*ormapi.Organization)
	for ii, _ := range orgs {
		orgsT[orgs[ii].Name] = &orgs[ii]
	}

	for name, org := range orgsT {
		name = util.GitlabGroupSanitize(name)
		if _, found := groupsT[name]; found {
			delete(groupsT, name)
		} else {
			// missing from gitlab, so create
			log.DebugLog(log.DebugLevelApi,
				"Gitlab Sync create missing group",
				"org", name)
			gitlabCreateGroup(org)
		}
	}
	for _, group := range groupsT {
		ca, _, err := gitlabClient.CustomAttribute.GetCustomGroupAttribute(group.ID, "createdby")
		if err != nil {
			continue
		}
		if ca.Value != "mastercontroller" {
			continue
		}
		// delete extra group created by master controller
		log.DebugLog(log.DebugLevelApi,
			"Gitlab Sync delete extra group",
			"name", group.Name)
		_, err = gitlabClient.Groups.DeleteGroup(group.ID)
		if err != nil {
			s.syncErr(err)
		}
	}
}

func (s *GitlabSync) syncGroupMembers() {
	members := make(map[string]map[string]*gitlab.GroupMember)
	var err error

	groupings := enforcer.GetGroupingPolicy()
	for ii, _ := range groupings {
		role := parseRole(groupings[ii])
		if role == nil || role.Org == "" {
			continue
		}
		// get cached group
		memberTable, found := members[role.Org]
		if !found {
			gname := util.GitlabGroupSanitize(role.Org)
			memberlist, _, err := gitlabClient.Groups.ListGroupMembers(gname, &gitlab.ListGroupMembersOptions{})
			if err != nil {
				s.syncErr(err)
				continue
			}
			// convert list to table for easier processing
			memberTable = make(map[string]*gitlab.GroupMember)
			for _, member := range memberlist {
				memberTable[member.Username] = member
			}
			members[role.Org] = memberTable
		}
		found = false
		for name, _ := range memberTable {
			if name == role.Username {
				found = true
				delete(memberTable, name)
				break
			}
		}
		if found {
			continue
		}
		// add member back to group
		log.DebugLog(log.DebugLevelApi,
			"Gitlab Sync restore role", "role", role)
		gitlabAddGroupMember(role)
	}
	// delete members that shouldn't be part of the group anymore
	for roleOrg, memberTable := range members {
		for _, groupMember := range memberTable {
			if groupMember.ID == 1 {
				// root is always member of a group
				continue
			}
			log.DebugLog(log.DebugLevelApi,
				"Gitlab Sync remove extra role",
				"org", roleOrg, "member", groupMember.Username)
			gname := util.GitlabGroupSanitize(roleOrg)
			_, err = gitlabClient.GroupMembers.RemoveGroupMember(gname, groupMember.ID)
			if err != nil {
				s.syncErr(err)
			}
		}
	}
}

func (s *GitlabSync) NeedsSync() {
	s.needsSync = true
}

func (s *GitlabSync) wakeup() {
	select {
	case s.run <- true:
	default:
	}
}

func (s *GitlabSync) syncErr(err error) {
	log.DebugLog(log.DebugLevelApi, "Gitlab Sync failed", "err", err)
	s.NeedsSync()
}

func GitlabResync(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	// only super user can access
	if !enforcer.Enforce(claims.Username, "", ResourceControllers, ActionManage) {
		return echo.ErrForbidden
	}
	gitlabSync.NeedsSync()
	gitlabSync.wakeup()
	return nil
}
