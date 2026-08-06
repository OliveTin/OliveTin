import { reactive, computed } from 'vue'
import {
  CellsIcon,
  DashboardSquare01Icon,
  LeftToRightListDashIcon,
  PlayIcon,
  Wrench01Icon
} from '@hugeicons/core-free-icons'

export const MAX_SEARCH_HINT_ACTIONS = 100
export const MAX_SEARCH_HINT_DASHBOARDS = 100
export const MAX_SEARCH_HINT_ENTITIES_PER_TYPE = 50

const SOURCE_ENTITIES = 'entities'
const SOURCE_ACTIONS = 'actions'
const SOURCE_DASHBOARDS = 'dashboards'
const SOURCE_NAVIGATION = 'navigation'

// QuickSearch keeps insertion order; prefer entities over actions for equal matches.
const CATEGORY_PRIORITY = {
  Navigation: 0,
  Dashboards: 1,
  Entities: 2,
  Actions: 3
}

const state = reactive({
  itemsById: {},
  sourceToIds: {}
})

export const searchIndexItems = computed(() => {
  return Object.values(state.itemsById).sort((a, b) => {
    return categoryPriority(a.category) - categoryPriority(b.category)
  })
})

function categoryPriority (category) {
  return CATEGORY_PRIORITY[category] ?? 50
}

export function clearSearchIndex () {
  state.itemsById = {}
  state.sourceToIds = {}
}

function replaceSource (source, items) {
  const previousIds = state.sourceToIds[source] || []

  for (const itemId of previousIds) {
    delete state.itemsById[itemId]
  }

  const nextIds = []
  for (const item of items) {
    if (!item?.id) {
      continue
    }

    state.itemsById[item.id] = item
    nextIds.push(item.id)
  }

  state.sourceToIds[source] = nextIds
}

export function dashboardRoutePath (title, entityType, entityKey) {
  if (!title) {
    return '/'
  }

  if (title === 'Actions' && !entityType && !entityKey) {
    return '/'
  }

  let path = `/dashboards/${title}`

  if (entityType && entityKey) {
    path += `/${entityType}/${entityKey}`
  }

  return path
}

/**
 * Indexes search hints from Init (entities and actions).
 * Dashboards are indexed separately from rootDashboardEntries.
 */
export function indexSearchHints (searchHints) {
  replaceSource(SOURCE_ENTITIES, entityItemsFromHints(searchHints?.entities))
  replaceSource(SOURCE_ACTIONS, actionItemsFromHints(searchHints?.actions))
}

/**
 * Indexes dashboards from Init.rootDashboardEntries (already ACL-filtered).
 */
export function indexRootDashboardEntries (entries) {
  replaceSource(SOURCE_DASHBOARDS, dashboardItemsFromRootEntries(entries))
}

function entityItemsFromHints (hints) {
  if (!Array.isArray(hints)) {
    return []
  }

  const countsByType = {}
  const items = []

  for (const hint of hints) {
    if (!hint?.uniqueKey || !hint?.type) {
      continue
    }

    const count = countsByType[hint.type] || 0
    if (count >= MAX_SEARCH_HINT_ENTITIES_PER_TYPE) {
      continue
    }

    countsByType[hint.type] = count + 1
    items.push({
      id: `entity:${hint.type}:${hint.uniqueKey}`,
      title: hint.title || hint.uniqueKey,
      description: hint.type,
      category: 'Entities',
      type: 'route',
      path: `/entity-details/${hint.type}/${hint.uniqueKey}`,
      icon: CellsIcon
    })
  }

  return items
}

function actionItemsFromHints (hints) {
  if (!Array.isArray(hints)) {
    return []
  }

  const items = []
  for (const hint of hints.slice(0, MAX_SEARCH_HINT_ACTIONS)) {
    if (!hint?.bindingId) {
      continue
    }

    items.push({
      id: `action:${hint.bindingId}`,
      title: hint.title || hint.bindingId,
      category: 'Actions',
      type: 'route',
      path: `/action/${hint.bindingId}`,
      icon: PlayIcon
    })
  }

  return items
}

function dashboardItemsFromRootEntries (entries) {
  if (!Array.isArray(entries)) {
    return []
  }

  const items = []
  for (const entry of entries.slice(0, MAX_SEARCH_HINT_DASHBOARDS)) {
    if (!entry?.title) {
      continue
    }

    const section = (entry.category || '').trim()

    items.push({
      id: `dashboard:${entry.title}`,
      title: entry.title,
      description: section,
      category: 'Dashboards',
      type: 'route',
      path: dashboardRoutePath(entry.title),
      icon: DashboardSquare01Icon
    })
  }

  return items
}

/**
 * Indexes system navigation destinations that the user can already reach from
 * the sidebar, respecting Init visibility flags.
 */
export function indexSystemNavigation ({
  showLogs = false,
  showDiagnostics = false,
  entitiesTitle = 'Entities',
  logsTitle = 'Logs',
  diagnosticsTitle = 'Diagnostics'
} = {}) {
  const items = [{
    id: 'nav:entities',
    title: entitiesTitle,
    category: 'Navigation',
    type: 'route',
    path: '/entities',
    icon: CellsIcon
  }]

  if (showLogs) {
    items.push({
      id: 'nav:logs',
      title: logsTitle,
      category: 'Navigation',
      type: 'route',
      path: '/logs',
      icon: LeftToRightListDashIcon
    })
  }

  if (showDiagnostics) {
    items.push({
      id: 'nav:diagnostics',
      title: diagnosticsTitle,
      category: 'Navigation',
      type: 'route',
      path: '/diagnostics',
      icon: Wrench01Icon
    })
  }

  replaceSource(SOURCE_NAVIGATION, items)
}
