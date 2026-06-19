package executor

import (
	"fmt"

	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/entities"
)

type DashboardNavigationTarget struct {
	Title      string
	EntityType string
	EntityKey  string
	Path       string
}

func (target DashboardNavigationTarget) key() string {
	return target.Title + "\x00" + target.EntityType + "\x00" + target.EntityKey
}

func (b *ActionBinding) IsOnConfiguredDashboard() bool {
	for _, dashboard := range b.OnDashboards {
		if dashboard.Title != "Actions" {
			return true
		}
	}
	return false
}

type dashboardTargetIndex struct {
	byTitle          map[string][]DashboardNavigationTarget
	byTitleEntityKey map[string]map[string][]DashboardNavigationTarget
}

func buildDashboardTargetIndex(cfg *config.Config) *dashboardTargetIndex {
	index := &dashboardTargetIndex{
		byTitle:          make(map[string][]DashboardNavigationTarget),
		byTitleEntityKey: make(map[string]map[string][]DashboardNavigationTarget),
	}

	for _, dashboard := range cfg.Dashboards {
		walkDashboardContents(dashboard.Contents, dashboard.Title, index)
	}

	return index
}

func walkDashboardContents(contents []*config.DashboardComponent, rootDashboardTitle string, index *dashboardTargetIndex) {
	for _, component := range contents {
		walkDashboardComponent(component, rootDashboardTitle, index)
	}
}

func walkDashboardComponent(component *config.DashboardComponent, rootDashboardTitle string, index *dashboardTargetIndex) {
	if component.Type == "fieldset" && component.Entity != "" {
		walkEntityFieldset(component, rootDashboardTitle, component.Entity, index)
		return
	}

	recordActionTarget(component, rootDashboardTitle, "", "", index)

	if len(component.Contents) > 0 {
		walkDashboardContents(component.Contents, rootDashboardTitle, index)
	}
}

func walkEntityFieldset(fieldset *config.DashboardComponent, rootDashboardTitle, entityType string, index *dashboardTargetIndex) {
	for _, component := range fieldset.Contents {
		if component.Type == "directory" {
			walkEntityDirectory(component, entityType, index)
			continue
		}

		recordActionTargetForAllEntities(component, rootDashboardTitle, entityType, index)

		if len(component.Contents) > 0 {
			walkEntityFieldsetContents(component.Contents, rootDashboardTitle, entityType, index)
		}
	}
}

func walkEntityFieldsetContents(contents []*config.DashboardComponent, rootDashboardTitle, entityType string, index *dashboardTargetIndex) {
	for _, component := range contents {
		if component.Type == "directory" {
			walkEntityDirectory(component, entityType, index)
			continue
		}

		recordActionTargetForAllEntities(component, rootDashboardTitle, entityType, index)

		if len(component.Contents) > 0 {
			walkEntityFieldsetContents(component.Contents, rootDashboardTitle, entityType, index)
		}
	}
}

func walkEntityDirectory(directory *config.DashboardComponent, entityType string, index *dashboardTargetIndex) {
	for _, entity := range entities.GetEntityInstancesOrdered(entityType) {
		for _, component := range directory.Contents {
			recordActionTarget(component, directory.Title, entityType, entity.UniqueKey, index)
		}
	}
}

func recordActionTargetForAllEntities(component *config.DashboardComponent, rootDashboardTitle, entityType string, index *dashboardTargetIndex) {
	actionTitle := actionTitleFromComponent(component)
	if actionTitle == "" {
		return
	}

	target := dashboardNavigationTarget(rootDashboardTitle, "", "")
	for _, entity := range entities.GetEntityInstancesOrdered(entityType) {
		addEntityTarget(index, actionTitle, entity.UniqueKey, target)
	}
}

func recordActionTarget(component *config.DashboardComponent, dashboardTitle, entityType, entityKey string, index *dashboardTargetIndex) {
	actionTitle := actionTitleFromComponent(component)
	if actionTitle == "" {
		return
	}

	target := dashboardNavigationTarget(dashboardTitle, entityType, entityKey)
	if entityType != "" && entityKey != "" {
		addEntityTarget(index, actionTitle, entityKey, target)
		return
	}

	addTitleTarget(index, actionTitle, target)
}

func actionTitleFromComponent(component *config.DashboardComponent) string {
	if title := inlineActionTitle(component); title != "" {
		return title
	}

	if component.Type == "link" || component.Type == "" {
		return component.Title
	}

	return ""
}

func inlineActionTitle(component *config.DashboardComponent) string {
	if component.InlineAction == nil {
		return ""
	}

	if component.Title != "" {
		return component.Title
	}

	return component.InlineAction.Title
}

func dashboardNavigationTarget(title, entityType, entityKey string) DashboardNavigationTarget {
	return DashboardNavigationTarget{
		Title:      title,
		EntityType: entityType,
		EntityKey:  entityKey,
		Path:       dashboardNavigationPath(title, entityType, entityKey),
	}
}

func dashboardNavigationPath(title, entityType, entityKey string) string {
	if title == "Actions" {
		return "/"
	}

	if entityType != "" && entityKey != "" {
		return fmt.Sprintf("/dashboards/%s/%s/%s", title, entityType, entityKey)
	}

	return fmt.Sprintf("/dashboards/%s", title)
}

func addTitleTarget(index *dashboardTargetIndex, actionTitle string, target DashboardNavigationTarget) {
	index.byTitle[actionTitle] = appendUniqueTarget(index.byTitle[actionTitle], target)
}

func addEntityTarget(index *dashboardTargetIndex, actionTitle, entityKey string, target DashboardNavigationTarget) {
	if index.byTitleEntityKey[actionTitle] == nil {
		index.byTitleEntityKey[actionTitle] = make(map[string][]DashboardNavigationTarget)
	}

	entityTargets := index.byTitleEntityKey[actionTitle][entityKey]
	index.byTitleEntityKey[actionTitle][entityKey] = appendUniqueTarget(entityTargets, target)
}

func appendUniqueTarget(targets []DashboardNavigationTarget, target DashboardNavigationTarget) []DashboardNavigationTarget {
	for _, existing := range targets {
		if existing.key() == target.key() {
			return targets
		}
	}

	return append(targets, target)
}

func (index *dashboardTargetIndex) targetsForAction(actionTitle string) []DashboardNavigationTarget {
	return dedupeTargets(index.byTitle[actionTitle])
}

func (index *dashboardTargetIndex) targetsForEntityAction(actionTitle, entityKey string) []DashboardNavigationTarget {
	targets := make([]DashboardNavigationTarget, 0)
	targets = append(targets, index.byTitle[actionTitle]...)

	if entityTargets, ok := index.byTitleEntityKey[actionTitle]; ok {
		targets = append(targets, entityTargets[entityKey]...)
	}

	return dedupeTargets(targets)
}

func dedupeTargets(targets []DashboardNavigationTarget) []DashboardNavigationTarget {
	if len(targets) == 0 {
		return nil
	}

	seen := make(map[string]bool, len(targets))
	result := make([]DashboardNavigationTarget, 0, len(targets))

	for _, target := range targets {
		key := target.key()
		if seen[key] {
			continue
		}

		seen[key] = true
		result = append(result, target)
	}

	return result
}

func resolveOnDashboards(index *dashboardTargetIndex, actionTitle, entityKey string) []DashboardNavigationTarget {
	var targets []DashboardNavigationTarget
	if entityKey == "" {
		targets = index.targetsForAction(actionTitle)
	} else {
		targets = index.targetsForEntityAction(actionTitle, entityKey)
	}

	if len(targets) == 0 {
		return []DashboardNavigationTarget{dashboardNavigationTarget("Actions", "", "")}
	}

	return targets
}
