package api

import (
	"sort"
	"strings"

	apiv1 "github.com/OliveTin/OliveTin/gen/olivetin/api/v1"
	authpublic "github.com/OliveTin/OliveTin/internal/auth/authpublic"
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/entities"
)

const (
	defaultEntityInstancesPageSize = 10
	maxEntityInstancesPageSize     = 100
)

func (api *oliveTinAPI) buildEntityDefinitionsResponse(user *authpublic.AuthenticatedUser, req *apiv1.GetEntitiesRequest, entityMap entities.EntitiesByClass) []*apiv1.EntityDefinition {
	if req != nil && req.EntityType != "" {
		return api.buildFilteredEntityDefinitions(user, req, entityMap)
	}

	return api.buildAllEntityDefinitions(user, entityMap)
}

func (api *oliveTinAPI) buildAllEntityDefinitions(user *authpublic.AuthenticatedUser, entityMap entities.EntitiesByClass) []*apiv1.EntityDefinition {
	entityNames := sortedEntityTypeNames(entityMap)
	entityDefinitions := make([]*apiv1.EntityDefinition, 0, len(entityNames))

	for _, name := range entityNames {
		if !api.userCanViewEntityType(user, name) {
			continue
		}

		def := api.buildEntityDefinition(name, entityMap[name], false, "", 0, 0)
		entityDefinitions = append(entityDefinitions, def)
	}

	return entityDefinitions
}

func (api *oliveTinAPI) buildFilteredEntityDefinitions(user *authpublic.AuthenticatedUser, req *apiv1.GetEntitiesRequest, entityMap entities.EntitiesByClass) []*apiv1.EntityDefinition {
	if !api.userCanViewEntityType(user, req.EntityType) {
		return nil
	}

	entityInstances, ok := entityMap[req.EntityType]
	if !ok || len(entityInstances) == 0 {
		return nil
	}

	pageSize := normalizeEntityInstancesPageSize(req.PageSize)
	page := normalizeEntityInstancesPage(req.Page)
	def := api.buildEntityDefinition(req.EntityType, entityInstances, true, req.Filter, page, pageSize)

	return []*apiv1.EntityDefinition{def}
}

func (api *oliveTinAPI) buildEntityDefinition(entityType string, entityInstances map[string]*entities.Entity, paginate bool, filter string, page, pageSize int32) *apiv1.EntityDefinition {
	entityFile := entityFileForType(api.cfg, entityType)
	properties := entityPropertiesFromFile(entityFile)
	instances := buildSortedEntityInstances(entityType, entityInstances, properties)

	def := &apiv1.EntityDefinition{
		Title:            entityType,
		UsedOnDashboards: findDashboardsForEntity(entityType, api.cfg.Dashboards),
		Icon:             entityTypeIcon(api.cfg, entityType),
		Properties:       entityDefinitionProperties(properties),
		TotalInstances:   int32(len(instances)),
	}

	if !paginate {
		if len(properties) == 0 {
			def.Instances = instances
		}
		return def
	}

	filtered := filterEntityInstances(instances, filter)
	def.TotalInstances = int32(len(filtered))
	def.Instances = paginateEntityInstances(filtered, page, pageSize)
	return def
}

func sortedEntityTypeNames(entityMap entities.EntitiesByClass) []string {
	entityNames := make([]string, 0, len(entityMap))
	for name := range entityMap {
		entityNames = append(entityNames, name)
	}
	sort.Strings(entityNames)
	return entityNames
}

func normalizeEntityInstancesPage(page int32) int32 {
	if page < 1 {
		return 1
	}
	return page
}

func normalizeEntityInstancesPageSize(pageSize int32) int32 {
	if pageSize < 1 {
		return defaultEntityInstancesPageSize
	}
	if pageSize > maxEntityInstancesPageSize {
		return maxEntityInstancesPageSize
	}
	return pageSize
}

func filterEntityInstances(instances []*apiv1.Entity, filter string) []*apiv1.Entity {
	filter = strings.TrimSpace(strings.ToLower(filter))
	if filter == "" {
		return instances
	}

	filtered := make([]*apiv1.Entity, 0, len(instances))
	for _, instance := range instances {
		if entityInstanceMatchesFilter(instance, filter) {
			filtered = append(filtered, instance)
		}
	}
	return filtered
}

func entityInstanceMatchesFilter(instance *apiv1.Entity, filter string) bool {
	if instance == nil {
		return false
	}

	if stringContainsFold(instance.Title, filter) || stringContainsFold(instance.UniqueKey, filter) {
		return true
	}

	return entityFieldsContainFilter(instance.Fields, filter)
}

func stringContainsFold(value, filter string) bool {
	return strings.Contains(strings.ToLower(value), strings.ToLower(filter))
}

func entityFieldsContainFilter(fields map[string]string, filter string) bool {
	for _, value := range fields {
		if stringContainsFold(value, filter) {
			return true
		}
	}

	return false
}

func paginateEntityInstances(instances []*apiv1.Entity, page, pageSize int32) []*apiv1.Entity {
	count := int64(len(instances))
	start := int64(page-1) * int64(pageSize)
	if start >= count {
		return []*apiv1.Entity{}
	}

	end := start + int64(pageSize)
	if end > count {
		end = count
	}

	return instances[int(start):int(end)]
}

func entityFieldsForResponse(data any, properties []config.EntityProperty) map[string]string {
	if len(properties) > 0 {
		return entityListFields(data, properties)
	}

	return serializeEntityFields(data)
}
