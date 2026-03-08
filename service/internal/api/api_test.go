package api

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	log "github.com/sirupsen/logrus"

	apiv1 "github.com/OliveTin/OliveTin/gen/olivetin/api/v1"
	apiv1connect "github.com/OliveTin/OliveTin/gen/olivetin/api/v1/apiv1connect"
	authpublic "github.com/OliveTin/OliveTin/internal/auth/authpublic"
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/entities"
	"github.com/OliveTin/OliveTin/internal/executor"

	"net/http"
	"net/http/httptest"
	"path"
)

func getNewTestServerAndClient(injectedConfig *config.Config) (*httptest.Server, apiv1connect.OliveTinApiServiceClient) {
	ex := executor.DefaultExecutor(injectedConfig)
	ex.RebuildActionMap()

	apiPath, apiHandler := GetNewHandler(ex)

	mux := http.NewServeMux()
	mux.Handle("/api/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Infof("HTTP Request: %s %s", r.Method, r.URL.Path)

		// Translate /api/<service>/<method> to <service>/<method>
		fn := path.Base(r.URL.Path)
		r.URL.Path = apiPath + fn

		apiHandler.ServeHTTP(w, r)
	}))

	log.Infof("API path is %s", apiPath)

	httpclient := &http.Client{}

	ts := httptest.NewServer(mux)

	client := apiv1connect.NewOliveTinApiServiceClient(httpclient, ts.URL+"/api")

	log.Infof("Test server URL is %s", ts.URL+"/api"+apiPath)

	return ts, client
}

func TestGetActionsAndStart(t *testing.T) {
	cfg := config.DefaultConfig()

	btn1 := &config.Action{}
	btn1.Title = "blat"
	btn1.ID = "blat"
	btn1.Shell = "echo 'test'"
	cfg.Actions = append(cfg.Actions, btn1)

	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()

	conn, client := getNewTestServerAndClient(cfg)

	respInit, errInit := client.Init(context.Background(), connect.NewRequest(&apiv1.InitRequest{}))
	respGetReady, errReady := client.GetReadyz(context.Background(), connect.NewRequest(&apiv1.GetReadyzRequest{}))

	if errInit != nil {
		t.Errorf("Init request failed: %v", errInit)
		return
	}

	if errReady != nil {
		t.Errorf("GetReadyz request failed: %v", errReady)
		return
	}

	log.Infof("GetReadyz response: %v", respGetReady.Msg)

	assert.Equal(t, true, true, "sayHello Failed")

	//	assert.Equal(t, 1, len(respGb.Msg.Actions), "Got 1 action button back")

	log.Printf("Response: %+v", respInit)

	respSa, err := client.StartAction(context.Background(), connect.NewRequest(&apiv1.StartActionRequest{
		//		ActionId: "blat"
	}))

	assert.NotNil(t, err, "Error 404 after start action")
	assert.Nil(t, respSa, "Nil response for non existing action")

	defer conn.Close()
}

func TestGetEntities(t *testing.T) {
	cfg := config.DefaultConfig()

	ts, client := getNewTestServerAndClient(cfg)
	defer ts.Close()

	setupTestEntities()

	resp, err := client.GetEntities(context.Background(), connect.NewRequest(&apiv1.GetEntitiesRequest{}))

	assert.NoError(t, err, "GetEntities should not return an error")
	assert.NotNil(t, resp, "GetEntities response should not be nil")
	assert.NotNil(t, resp.Msg, "GetEntities response message should not be nil")

	entityDefinitions := resp.Msg.EntityDefinitions
	assert.Equal(t, 3, len(entityDefinitions), "Should return 3 entity definitions")

	validateEntityOrderAndStructure(t, entityDefinitions)
	validateNoDuplicates(t, entityDefinitions)
	validateConsistency(t, client, entityDefinitions)
}

func setupTestEntities() {
	entities.ClearEntitiesOfType("server")
	entities.ClearEntitiesOfType("database")
	entities.ClearEntitiesOfType("application")

	entities.AddEntity("server", "zebra", map[string]any{"title": "Server Zebra", "hostname": "zebra.example.com"})
	entities.AddEntity("server", "alpha", map[string]any{"title": "Server Alpha", "hostname": "alpha.example.com"})
	entities.AddEntity("server", "beta", map[string]any{"title": "Server Beta", "hostname": "beta.example.com"})

	entities.AddEntity("database", "mysql", map[string]any{"title": "MySQL Database", "type": "mysql"})
	entities.AddEntity("database", "postgres", map[string]any{"title": "PostgreSQL Database", "type": "postgres"})

	entities.AddEntity("application", "webapp", map[string]any{"title": "Web Application", "port": 8080})
}

func validateEntityOrderAndStructure(t *testing.T, entityDefinitions []*apiv1.EntityDefinition) {
	assert.Equal(t, "application", entityDefinitions[0].Title, "First entity should be 'application' (alphabetically first)")
	assert.Equal(t, 1, len(entityDefinitions[0].Instances), "Application should have 1 instance")
	assert.Equal(t, "webapp", entityDefinitions[0].Instances[0].UniqueKey, "Application instance should be 'webapp'")

	assert.Equal(t, "database", entityDefinitions[1].Title, "Second entity should be 'database' (alphabetically second)")
	assert.Equal(t, 2, len(entityDefinitions[1].Instances), "Database should have 2 instances")
	assert.Equal(t, "mysql", entityDefinitions[1].Instances[0].UniqueKey, "First database instance should be 'mysql' (alphabetically first)")
	assert.Equal(t, "postgres", entityDefinitions[1].Instances[1].UniqueKey, "Second database instance should be 'postgres' (alphabetically second)")

	assert.Equal(t, "server", entityDefinitions[2].Title, "Third entity should be 'server' (alphabetically third)")
	assert.Equal(t, 3, len(entityDefinitions[2].Instances), "Server should have 3 instances")
	assert.Equal(t, "alpha", entityDefinitions[2].Instances[0].UniqueKey, "First server instance should be 'alpha' (alphabetically first)")
	assert.Equal(t, "beta", entityDefinitions[2].Instances[1].UniqueKey, "Second server instance should be 'beta' (alphabetically second)")
	assert.Equal(t, "zebra", entityDefinitions[2].Instances[2].UniqueKey, "Third server instance should be 'zebra' (alphabetically third)")
}

func validateNoDuplicates(t *testing.T, entityDefinitions []*apiv1.EntityDefinition) {
	instanceKeys := make(map[string]map[string]bool)
	for _, def := range entityDefinitions {
		instanceKeys[def.Title] = make(map[string]bool)
		for _, inst := range def.Instances {
			assert.False(t, instanceKeys[def.Title][inst.UniqueKey], "Instance key %s should not be duplicated in entity %s", inst.UniqueKey, def.Title)
			instanceKeys[def.Title][inst.UniqueKey] = true
		}
	}
}

func validateConsistency(t *testing.T, client apiv1connect.OliveTinApiServiceClient, entityDefinitions []*apiv1.EntityDefinition) {
	resp2, err2 := client.GetEntities(context.Background(), connect.NewRequest(&apiv1.GetEntitiesRequest{}))
	assert.NoError(t, err2, "Second GetEntities call should not return an error")
	assert.Equal(t, len(entityDefinitions), len(resp2.Msg.EntityDefinitions), "Second call should return same number of entity definitions")

	for i, def := range entityDefinitions {
		assert.Equal(t, def.Title, resp2.Msg.EntityDefinitions[i].Title, "Entity order should be consistent across calls")
		assert.Equal(t, len(def.Instances), len(resp2.Msg.EntityDefinitions[i].Instances), "Instance count should be consistent")
		for j, inst := range def.Instances {
			assert.Equal(t, inst.UniqueKey, resp2.Msg.EntityDefinitions[i].Instances[j].UniqueKey, "Instance order should be consistent across calls")
		}
	}
}

func TestEvaluateEnabledExpression(t *testing.T) {
	tests := []struct {
		name           string
		expression     string
		entity         *entities.Entity
		expectedResult bool
	}{
		{
			name:           "empty expression returns true",
			expression:     "",
			entity:         nil,
			expectedResult: true,
		},
		{
			name:           "literal true returns true",
			expression:     "true",
			entity:         nil,
			expectedResult: true,
		},
		{
			name:           "literal True returns true (case insensitive)",
			expression:     "True",
			entity:         nil,
			expectedResult: true,
		},
		{
			name:           "literal 1 returns true",
			expression:     "1",
			entity:         nil,
			expectedResult: true,
		},
		{
			name:           "literal false returns false",
			expression:     "false",
			entity:         nil,
			expectedResult: false,
		},
		{
			name:           "literal 0 returns false",
			expression:     "0",
			entity:         nil,
			expectedResult: false,
		},
		{
			name:           "empty result returns false",
			expression:     "{{ .NonExistent }}",
			entity:         nil,
			expectedResult: false,
		},
		{
			name:           "expression with CurrentEntity true",
			expression:     "{{ eq .CurrentEntity.powered_on true }}",
			entity:         &entities.Entity{Data: map[string]any{"powered_on": true}},
			expectedResult: true,
		},
		{
			name:           "expression with CurrentEntity false",
			expression:     "{{ eq .CurrentEntity.powered_on true }}",
			entity:         &entities.Entity{Data: map[string]any{"powered_on": false}},
			expectedResult: false,
		},
		{
			name:           "expression with CurrentEntity integer 1",
			expression:     "{{ .CurrentEntity.status }}",
			entity:         &entities.Entity{Data: map[string]any{"status": 1}},
			expectedResult: true,
		},
		{
			name:           "expression with CurrentEntity integer 0",
			expression:     "{{ .CurrentEntity.status }}",
			entity:         &entities.Entity{Data: map[string]any{"status": 0}},
			expectedResult: false,
		},
		{
			name:           "template parse error returns false",
			expression:     "{{ invalid syntax }}",
			entity:         nil,
			expectedResult: false,
		},
		{
			name:           "template exec error returns false",
			expression:     "{{ .CurrentEntity.nonexistent }}",
			entity:         nil,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := &config.Action{
				EnabledExpression: tt.expression,
			}
			result := evaluateEnabledExpression(action, tt.entity)
			assert.Equal(t, tt.expectedResult, result, "evaluateEnabledExpression should return expected result")
		})
	}
}

func TestBuildActionWithEnabledExpression(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.DefaultPermissions.Exec = true

	action := &config.Action{
		Title:             "Test Action",
		Shell:             "echo test",
		EnabledExpression: "{{ eq .CurrentEntity.enabled true }}",
	}
	cfg.Actions = append(cfg.Actions, action)

	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()

	binding := findBindingByTitle(ex, "Test Action")
	assert.NotNil(t, binding, "Binding should be found")

	rr := &DashboardRenderRequest{
		AuthenticatedUser: &authpublic.AuthenticatedUser{Username: "testuser"},
		cfg:               cfg,
		ex:                ex,
	}

	testWithEntity(t, binding, rr, true, true, "Action should be executable when entity.enabled is true")
	testWithEntity(t, binding, rr, false, false, "Action should not be executable when entity.enabled is false")

	bindingNoExpr := findBindingByTitle(ex, "Test Action No Expression")
	if bindingNoExpr == nil {
		actionNoExpression := &config.Action{
			Title: "Test Action No Expression",
			Shell: "echo test",
		}
		cfg.Actions = append(cfg.Actions, actionNoExpression)
		ex.RebuildActionMap()
		bindingNoExpr = findBindingByTitle(ex, "Test Action No Expression")
	}

	actionResult := buildAction(bindingNoExpr, rr)
	assert.True(t, actionResult.CanExec, "Action without enabledExpression should be executable")
}

func findBindingByTitle(ex *executor.Executor, title string) *executor.ActionBinding {
	ex.MapActionBindingsLock.RLock()
	defer ex.MapActionBindingsLock.RUnlock()

	for _, b := range ex.MapActionBindings {
		if b.Action.Title == title {
			return b
		}
	}
	return nil
}

func testWithEntity(t *testing.T, binding *executor.ActionBinding, rr *DashboardRenderRequest, enabled bool, expectedCanExec bool, message string) {
	binding.Entity = &entities.Entity{
		UniqueKey: "test-entity",
		Data:      map[string]any{"enabled": enabled},
	}

	actionResult := buildAction(binding, rr)
	assert.Equal(t, expectedCanExec, actionResult.CanExec, message)
}

// buildViewPermissionTestConfig returns config and users for GHSA view-permission tests:
// one action "secret_action", ACL "restricted" (view:false) for user "low", ACL "full" (view:true) for user "admin".
func buildViewPermissionTestConfig(t *testing.T) (*config.Config, *authpublic.AuthenticatedUser, *authpublic.AuthenticatedUser) {
	t.Helper()
	cfg := config.DefaultConfig()
	cfg.DefaultPermissions.View = false
	cfg.DefaultPermissions.Exec = false

	cfg.Actions = append(cfg.Actions, &config.Action{
		ID:    "secret_action",
		Title: "Secret Action",
		Shell: "echo sensitive",
		Icon:  "🔒",
	})

	cfg.AccessControlLists = append(cfg.AccessControlLists,
		&config.AccessControlList{
			Name:             "restricted",
			MatchUsernames:   []string{"low"},
			AddToEveryAction: true,
			Permissions:      config.PermissionsList{View: false, Exec: false, Logs: false, Kill: false},
		},
		&config.AccessControlList{
			Name:             "full",
			MatchUsernames:   []string{"admin"},
			AddToEveryAction: true,
			Permissions:      config.PermissionsList{View: true, Exec: true, Logs: true, Kill: true},
		},
	)

	lowUser := &authpublic.AuthenticatedUser{Username: "low"}
	lowUser.BuildUserAcls(cfg)
	adminUser := &authpublic.AuthenticatedUser{Username: "admin"}
	adminUser.BuildUserAcls(cfg)
	return cfg, lowUser, adminUser
}

// TestViewPermissionExcludedFromDashboard (GHSA: view permission) asserts that when a user has view: false,
// the default dashboard must not include that action. Covers GetDashboard not leaking action metadata.
func TestViewPermissionExcludedFromDashboard(t *testing.T) {
	cfg, lowUser, _ := buildViewPermissionTestConfig(t)
	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()

	rr := &DashboardRenderRequest{
		AuthenticatedUser: lowUser,
		cfg:               cfg,
		ex:                ex,
	}
	db := buildDefaultDashboard(rr)

	bindingIdsInDashboard := bindingIdsInDashboardContents(db.Contents)
	assert.NotContains(t, bindingIdsInDashboard, "secret_action",
		"user with view:false must not see action in dashboard; got bindingIds: %v", bindingIdsInDashboard)
}

// TestGetActionBindingDeniedWhenNoViewPermission (GHSA: view permission) asserts that GetActionBinding
// returns permission denied for a user with view: false. Covers GetActionBinding not exposing action details.
func TestGetActionBindingDeniedWhenNoViewPermission(t *testing.T) {
	cfg, lowUser, _ := buildViewPermissionTestConfig(t)
	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()
	api := newServer(ex)

	_, err := api.getActionBindingResponse(lowUser, "secret_action")
	require.Error(t, err)
	assert.Equal(t, connect.CodePermissionDenied, connect.CodeOf(err),
		"user with view:false must get permission denied from GetActionBinding")
}

// TestViewPermissionAllowedSeesAction (GHSA: view permission) asserts that a user with view: true
// still sees the action in the dashboard and can fetch it via GetActionBinding.
func TestViewPermissionAllowedSeesAction(t *testing.T) {
	cfg, _, adminUser := buildViewPermissionTestConfig(t)
	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()
	api := newServer(ex)

	rr := &DashboardRenderRequest{
		AuthenticatedUser: adminUser,
		cfg:               cfg,
		ex:                ex,
	}
	db := buildDefaultDashboard(rr)
	bindingIdsInDashboard := bindingIdsInDashboardContents(db.Contents)
	assert.Contains(t, bindingIdsInDashboard, "secret_action",
		"user with view:true must see action in dashboard; got bindingIds: %v", bindingIdsInDashboard)

	resp, err := api.getActionBindingResponse(adminUser, "secret_action")
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Action)
	assert.Equal(t, "secret_action", resp.Action.BindingId)
}

// TestViewPermissionExcludedFromCustomDashboard (issue #921) asserts that when a custom dashboard
// lists an action by title, users without view permission do not see that action (title or icon).
func TestViewPermissionExcludedFromCustomDashboard(t *testing.T) {
	cfg, lowUser, _ := buildViewPermissionTestConfig(t)
	cfg.Dashboards = []*config.DashboardComponent{
		{
			Title: "Custom",
			Contents: []*config.DashboardComponent{
				{Title: "Secret Action"},
			},
		},
	}
	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()

	rr := &DashboardRenderRequest{
		AuthenticatedUser: lowUser,
		cfg:               cfg,
		ex:                ex,
	}
	dashboard := findDashboardByTitle(rr, "Custom")
	require.NotNil(t, dashboard)
	db := buildDashboardFromConfig(dashboard, rr)
	require.NotNil(t, db)

	bindingIdsInDashboard := bindingIdsInDashboardContents(db.Contents)
	assert.NotContains(t, bindingIdsInDashboard, "secret_action",
		"user with view:false must not see action on custom dashboard; got bindingIds: %v", bindingIdsInDashboard)
}

// TestViewPermissionExcludedFromEntityDashboard (GHSA: view permission) asserts that when a dashboard
// has an entity fieldset listing an action, users without view permission do not see that action.
func TestViewPermissionExcludedFromEntityDashboard(t *testing.T) {
	entities.ClearEntitiesOfType("vp_entity_test")
	defer entities.ClearEntitiesOfType("vp_entity_test")
	entities.AddEntity("vp_entity_test", "1", map[string]any{"title": "Test Entity"})

	cfg, lowUser, _ := buildViewPermissionTestConfig(t)
	cfg.Dashboards = []*config.DashboardComponent{
		{
			Title: "WithEntity",
			Contents: []*config.DashboardComponent{
				{
					Title: "Servers", Type: "fieldset", Entity: "vp_entity_test",
					Contents: []*config.DashboardComponent{{Title: "Secret Action"}},
				},
			},
		},
	}
	ex := executor.DefaultExecutor(cfg)
	ex.RebuildActionMap()

	rr := &DashboardRenderRequest{
		AuthenticatedUser: lowUser,
		cfg:               cfg,
		ex:                ex,
	}
	dashboard := findDashboardByTitle(rr, "WithEntity")
	require.NotNil(t, dashboard)
	db := buildDashboardFromConfig(dashboard, rr)
	require.NotNil(t, db)

	bindingIdsInDashboard := bindingIdsInDashboardContents(db.Contents)
	assert.NotContains(t, bindingIdsInDashboard, "secret_action",
		"user with view:false must not see action in entity fieldset; got bindingIds: %v", bindingIdsInDashboard)
}

func bindingIdsInDashboardContents(contents []*apiv1.DashboardComponent) []string {
	var ids []string
	for _, c := range contents {
		ids = append(ids, bindingIdsFromComponent(c)...)
	}
	return ids
}

func bindingIdsFromComponent(c *apiv1.DashboardComponent) []string {
	if c == nil {
		return nil
	}
	var ids []string
	if c.Action != nil && c.Action.BindingId != "" {
		ids = append(ids, c.Action.BindingId)
	}
	return append(ids, bindingIdsInDashboardContents(c.Contents)...)
}

func TestOrderTopLevelDashboardComponents_RegularFieldsetsPreserveConfigOrder(t *testing.T) {
	zebra := &apiv1.DashboardComponent{Title: "Zebra", Type: "fieldset", EntityType: ""}
	alpha := &apiv1.DashboardComponent{Title: "Alpha", Type: "fieldset", EntityType: ""}
	root := &apiv1.DashboardComponent{Title: "Actions", Type: "fieldset", EntityType: ""}
	components := []*apiv1.DashboardComponent{zebra, alpha, root}

	out := orderTopLevelDashboardComponents(components)

	require.Len(t, out, 3)
	assert.Same(t, zebra, out[0], "first must be Zebra (config order)")
	assert.Same(t, alpha, out[1], "second must be Alpha (config order)")
	assert.Same(t, root, out[2], "third must be root Actions fieldset")
}

func TestOrderTopLevelDashboardComponents_SortablesSorted(t *testing.T) {
	entityBeta := &apiv1.DashboardComponent{Title: "Beta", Type: "fieldset", EntityType: "server"}
	entityAlpha := &apiv1.DashboardComponent{Title: "Alpha", Type: "fieldset", EntityType: "server"}
	components := []*apiv1.DashboardComponent{entityBeta, entityAlpha}

	out := orderTopLevelDashboardComponents(components)

	require.Len(t, out, 2)
	assert.Equal(t, "Alpha", out[0].Title, "sortables ordered by title")
	assert.Equal(t, "Beta", out[1].Title)
}
