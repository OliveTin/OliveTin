package api

import (
	"sort"
	"strings"

	apiv1 "github.com/OliveTin/OliveTin/gen/olivetin/api/v1"
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/entities"
)

const (
	defaultEntityInstancesPageSize = 10
	maxEntityInstancesPageSize     = 100
)

func (api *oliveTinAPI) buildEntityDefinitionsResponse(req *apiv1.GetEntitiesRequest, entityMap entities.EntitiesByClass) []*apiv1.EntityDefinition {
	if req != nil && req.EntityType != "" {
		return api.buildFilteredEntityDefinitions(req, entityMap)
	}

	return api.buildAllEntityDefinitions(entityMap)
}

func (api *oliveTinAPI) buildAllEntityDefinitions(entityMap entities.EntitiesByClass) []*apiv1.EntityDefinition {
	entityNames := sortedEntityTypeNames(entityMap)
	entityDefinitions := make([]*apiv1.EntityDefinition, 0, len(entityNames))

	for _, name := range entityNames {
		entityFile := entityFileForType(api.cfg, name)
		properties := entityPropertiesFromFile(entityFile)
		instances := buildSortedEntityInstances(name, entityMap[name], properties)

		def := &apiv1.EntityDefinition{
			Title:            name,
			UsedOnDashboards: findDashboardsForEntity(name, api.cfg.Dashboards),
			Icon:             entityTypeIcon(api.cfg, name),
			Properties:       entityDefinitionProperties(properties),
			TotalInstances:   int32(len(instances)),
		}

		if len(properties) == 0 {
			def.Instances = instances
		}

		entityDefinitions = append(entityDefinitions, def)
	}

	return entityDefinitions
}

func (api *oliveTinAPI) buildFilteredEntityDefinitions(req *apiv1.GetEntitiesRequest, entityMap entities.EntitiesByClass) []*apiv1.EntityDefinition {
	entityInstances, ok := entityMap[req.EntityType]
	if !ok || len(entityInstances) == 0 {
		return nil
	}

	entityFile := entityFileForType(api.cfg, req.EntityType)
	properties := entityPropertiesFromFile(entityFile)
	instances := buildSortedEntityInstances(req.EntityType, entityInstances, properties)
	filtered := filterEntityInstances(instances, req.Filter)
	pageSize := normalizeEntityInstancesPageSize(req.PageSize)
	page := normalizeEntityInstancesPage(req.Page)

	def := &apiv1.EntityDefinition{
		Title:            req.EntityType,
		UsedOnDashboards: findDashboardsForEntity(req.EntityType, api.cfg.Dashboards),
		Icon:             entityTypeIcon(api.cfg, req.EntityType),
		Properties:       entityDefinitionProperties(properties),
		TotalInstances:   int32(len(filtered)),
		Instances:        paginateEntityInstances(filtered, page, pageSize),
	}

	return []*apiv1.EntityDefinition{def}
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
	if strings.Contains(strings.ToLower(instance.Title), filter) {
		return true
	}

	for _, value := range instance.Fields {
		if strings.Contains(strings.ToLower(value), filter) {
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
