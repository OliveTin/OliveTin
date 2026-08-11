import test from 'node:test'
import assert from 'node:assert/strict'

import {
  MAX_SEARCH_HINT_ACTIONS,
  MAX_SEARCH_HINT_DASHBOARDS,
  MAX_SEARCH_HINT_ENTITIES_PER_TYPE,
  clearSearchIndex,
  dashboardRoutePath,
  indexRootDashboardEntries,
  indexSearchHints,
  indexSystemNavigation,
  searchIndexItems
} from '../stores/searchIndex.js'

function resetIndex () {
  clearSearchIndex()
}

test('dashboardRoutePath maps Actions to root', () => {
  assert.equal(dashboardRoutePath('Actions'), '/')
})

test('dashboardRoutePath builds dashboard and entity paths', () => {
  assert.equal(dashboardRoutePath('Servers'), '/dashboards/Servers')
  assert.equal(
    dashboardRoutePath('Servers', 'host', 'web01'),
    '/dashboards/Servers/host/web01'
  )
  assert.equal(
    dashboardRoutePath('Ops/Prod', 'host/type', 'key?1'),
    '/dashboards/Ops%2FProd/host%2Ftype/key%3F1'
  )
  assert.equal(
    dashboardRoutePath('Servers', 'host', 'key#frag'),
    '/dashboards/Servers/host/key%23frag'
  )
})

test('indexSearchHints indexes entities and actions', () => {
  resetIndex()

  indexSearchHints({
    entities: [
      { title: 'web01', type: 'host', uniqueKey: '0' },
      { title: '', type: 'host', uniqueKey: '1' },
      { title: 'skip', type: '', uniqueKey: 'x' },
      { title: 'slash host', type: 'host/type', uniqueKey: 'key?1' },
      { title: 'hash host', type: 'host', uniqueKey: 'key#frag' }
    ],
    actions: [
      { title: 'Ping Host', bindingId: 'bind-ping' },
      { title: '', bindingId: 'bind-empty-title' },
      { title: 'Ignored', bindingId: '' },
      { title: 'Slash Action', bindingId: 'bind/slash' },
      { title: 'Query Action', bindingId: 'bind?query' },
      { title: 'Hash Action', bindingId: 'bind#hash' }
    ]
  })

  const byId = Object.fromEntries(searchIndexItems.value.map((item) => [item.id, item]))

  assert.equal(byId['entity:host:0'].title, 'web01')
  assert.equal(byId['entity:host:0'].path, '/entity-details/host/0')
  assert.equal(byId['entity:host:0'].description, 'host')
  assert.equal(byId['entity:host:1'].title, '1')
  assert.equal(
    byId['entity:host/type:key?1'].path,
    '/entity-details/host%2Ftype/key%3F1'
  )
  assert.equal(
    byId['entity:host:key#frag'].path,
    '/entity-details/host/key%23frag'
  )

  assert.equal(byId['action:bind-ping'].title, 'Ping Host')
  assert.equal(byId['action:bind-ping'].path, '/action/bind-ping')
  assert.equal(byId['action:bind/slash'].path, '/action/bind%2Fslash')
  assert.equal(byId['action:bind?query'].path, '/action/bind%3Fquery')
  assert.equal(byId['action:bind#hash'].path, '/action/bind%23hash')
  assert.equal(byId['action:bind-ping'].category, 'Actions')
  assert.equal(byId['action:bind-empty-title'].title, 'bind-empty-title')

  assert.equal(Object.keys(byId).includes('entity:host:x'), false)
  assert.equal(Object.keys(byId).includes('action:'), false)
})

test('indexRootDashboardEntries indexes ACL-filtered dashboards', () => {
  resetIndex()

  indexRootDashboardEntries([
    { title: 'Actions', category: '' },
    { title: 'My Server', category: 'Infrastructure' },
    { title: '', category: 'Ignored' }
  ])

  const byId = Object.fromEntries(searchIndexItems.value.map((item) => [item.id, item]))
  assert.equal(byId['dashboard:Actions'].path, '/')
  assert.equal(byId['dashboard:My Server'].path, '/dashboards/My%20Server')
  assert.equal(byId['dashboard:My Server'].description, 'Infrastructure')
  assert.equal(byId['dashboard:My Server'].category, 'Dashboards')
})

test('indexSearchHints replaces prior entity and action snapshots', () => {
  resetIndex()

  indexSearchHints({
    entities: [{ title: 'old', type: 'host', uniqueKey: 'old' }],
    actions: [{ title: 'Old Action', bindingId: 'old-action' }]
  })
  indexRootDashboardEntries([{ title: 'Old Board', category: '' }])

  indexSearchHints({
    entities: [{ title: 'new', type: 'host', uniqueKey: 'new' }],
    actions: [{ title: 'New Action', bindingId: 'new-action' }]
  })
  indexRootDashboardEntries([{ title: 'New Board', category: 'Ops' }])

  const byId = Object.fromEntries(searchIndexItems.value.map((item) => [item.id, item]))
  assert.equal(byId['entity:host:old'], undefined)
  assert.equal(byId['action:old-action'], undefined)
  assert.equal(byId['dashboard:Old Board'], undefined)
  assert.equal(byId['entity:host:new'].title, 'new')
  assert.equal(byId['action:new-action'].title, 'New Action')
  assert.equal(byId['dashboard:New Board'].description, 'Ops')
})

test('caps actions globally, entities per type, and dashboards from root entries', () => {
  resetIndex()

  const actions = []
  for (let i = 0; i < MAX_SEARCH_HINT_ACTIONS + 20; i++) {
    actions.push({ title: `Action ${i}`, bindingId: `a-${i}` })
  }

  const entities = []
  for (let i = 0; i < MAX_SEARCH_HINT_ENTITIES_PER_TYPE + 20; i++) {
    entities.push({ title: `Host ${i}`, type: 'host', uniqueKey: `${i}` })
    entities.push({ title: `Container ${i}`, type: 'container', uniqueKey: `${i}` })
  }

  const dashboards = []
  for (let i = 0; i < MAX_SEARCH_HINT_DASHBOARDS + 20; i++) {
    dashboards.push({ title: `Board ${i}`, category: '' })
  }

  indexSearchHints({ actions, entities })
  indexRootDashboardEntries(dashboards)

  const items = searchIndexItems.value
  assert.equal(items.filter((item) => item.category === 'Actions').length, MAX_SEARCH_HINT_ACTIONS)
  assert.equal(items.filter((item) => item.category === 'Dashboards').length, MAX_SEARCH_HINT_DASHBOARDS)
  assert.equal(
    items.filter((item) => item.category === 'Entities' && item.description === 'host').length,
    MAX_SEARCH_HINT_ENTITIES_PER_TYPE
  )
  assert.equal(
    items.filter((item) => item.category === 'Entities' && item.description === 'container').length,
    MAX_SEARCH_HINT_ENTITIES_PER_TYPE
  )
})

test('indexSystemNavigation always includes entities and respects visibility flags', () => {
  resetIndex()

  indexSystemNavigation({
    showLogs: true,
    showDiagnostics: false,
    entitiesTitle: 'Entities',
    logsTitle: 'Logs',
    diagnosticsTitle: 'Diagnostics'
  })

  let byId = Object.fromEntries(searchIndexItems.value.map((item) => [item.id, item]))
  assert.equal(byId['nav:entities'].path, '/entities')
  assert.equal(byId['nav:logs'].path, '/logs')
  assert.equal(byId['nav:diagnostics'], undefined)

  indexSystemNavigation({
    showLogs: false,
    showDiagnostics: true
  })

  byId = Object.fromEntries(searchIndexItems.value.map((item) => [item.id, item]))
  assert.equal(byId['nav:entities'].path, '/entities')
  assert.equal(byId['nav:logs'], undefined)
  assert.equal(byId['nav:diagnostics'].path, '/diagnostics')
})

test('entities appear before actions in searchIndexItems', () => {
  resetIndex()

  indexSearchHints({
    entities: [{ title: 'server1', type: 'host', uniqueKey: '0' }],
    actions: [{ title: 'server1', bindingId: 'bind-server1' }]
  })

  const titlesByCategory = searchIndexItems.value.map((item) => [item.category, item.title])
  const entityIndex = titlesByCategory.findIndex(([category]) => category === 'Entities')
  const actionIndex = titlesByCategory.findIndex(([category]) => category === 'Actions')

  assert.ok(entityIndex >= 0)
  assert.ok(actionIndex >= 0)
  assert.ok(entityIndex < actionIndex)
})

test('search index has no Logs category', () => {
  resetIndex()

  indexSearchHints({
    entities: [{ title: 'web01', type: 'host', uniqueKey: '0' }],
    actions: [{ title: 'Backup', bindingId: 'bind-backup' }]
  })
  indexRootDashboardEntries([{ title: 'Ops', category: '' }])
  indexSystemNavigation({ showLogs: true, showDiagnostics: true })

  assert.equal(
    searchIndexItems.value.some((item) => item.category === 'Logs'),
    false
  )
})
