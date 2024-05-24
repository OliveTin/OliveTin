import './ActionButton.js' // To define action-button
import { ExecutionDialog } from './ExecutionDialog.js'

/**
 * This is a weird function that just sets some globals.
 */
export function initMarshaller () {
  window.changeDirectory = changeDirectory
  window.showSection = showSection

  window.executionDialog = new ExecutionDialog()

  window.logEntries = {}

  window.initialHash = window.location.hash

  window.currentSection = ''

  window.addEventListener('EventExecutionFinished', onExecutionFinished)
}

export function marshalDashboardComponentsJsonToHtml (json) {
  marshalActionsJsonToHtml(json)
  marshalDashboardStructureToHtml(json)

  document.getElementById('username').innerText = json.authenticatedUser

  changeDirectory(null)

  document.body.setAttribute('initial-marshal-complete', 'true')
}

function marshalActionsJsonToHtml (json) {
  const currentIterationTimestamp = Date.now()

  window.actionButtons = {}

  for (const jsonButton of json.actions) {
    let htmlButton = window.actionButtons[jsonButton.id]

    if (typeof htmlButton === 'undefined') {
      htmlButton = document.createElement('action-button')
      htmlButton.constructFromJson(jsonButton)

      window.actionButtons[jsonButton.title] = htmlButton
    }

    htmlButton.updateFromJson(jsonButton)
    htmlButton.updateIterationTimestamp = currentIterationTimestamp
  }

  // Remove existing, but stale buttons (that were not updated in this round)
  for (const existingButton of document.querySelectorAll('action-button')) {
    if (existingButton.updateIterationTimestamp !== currentIterationTimestamp) {
      existingButton.remove()
    }
  }
}

function onExecutionFinished (evt) {
  const logEntry = evt.payload.logEntry

  const actionButton = window.actionButtons[logEntry.actionTitle]

  if (actionButton === undefined) {
    return
  }

  switch (actionButton.popupOnStart) {
    case 'execution-button':
      document.querySelector('execution-button#execution-' + logEntry.executionTrackingId).onExecutionFinished(logEntry)
      break
    case 'execution-dialog-stdout-only':
    case 'execution-dialog':
      actionButton.onExecutionFinished(logEntry)

      // We don't need to fetch the logEntry for the dialog because we already
      // have it, so we open the dialog and it will get updated below.

      window.executionDialog.show()
      window.executionDialog.executionUuid = logEntry.uuid

      break
    default:
      actionButton.onExecutionFinished(logEntry)
      break
  }

  marshalLogsJsonToHtml({
    logs: [logEntry]
  })

  // If the current execution dialog is open, update that too
  if (window.executionDialog.dlg.open && window.executionDialog.executionUuid === logEntry.uuid) {
    window.executionDialog.renderExecutionResult({
      logEntry: logEntry
    })
  }
}

function showSection (title) {
  title = title.replaceAll(' ', '')

  window.currentSection = title

  for (const section of document.querySelectorAll('section')) {
    if (section.title === title) {
      section.style.display = 'block'
    } else {
      section.style.display = 'none'
    }
  }

  setSectionNavigationVisible(false)

  changeDirectory(null)
}

function setSectionNavigationVisible (visible) {
  const nav = document.querySelector('nav')
  const btn = document.getElementById('sidebar-toggler-button')

  if (document.body.classList.contains('has-sidebar')) {
    if (visible) {
      btn.setAttribute('aria-pressed', false)
      btn.setAttribute('aria-label', 'Open sidebar navigation')
      btn.innerHTML = '&laquo;'

      nav.classList.add('shown')
      nav.style.display = 'flex'
    } else {
      btn.setAttribute('aria-pressed', true)
      btn.setAttribute('aria-label', 'Close sidebar navigation')
      btn.innerHTML = '&#9776;'

      nav.classList.remove('shown')
      setTimeout(() => {
        nav.style.display = 'none'
      }, 600)
    }
  } else {
    btn.disabled = true
  }
}

export function setupSectionNavigation (style) {
  const nav = document.querySelector('nav')
  const btn = document.getElementById('sidebar-toggler-button')

  if (style === 'sidebar') {
    nav.classList.add('sidebar')

    document.body.classList.add('has-sidebar')

    btn.onclick = () => {
      if (nav.classList.contains('shown')) {
        setSectionNavigationVisible(false)
      } else {
        setSectionNavigationVisible(true)
      }
    }
  } else {
    nav.classList.add('topbar')

    document.body.classList.add('has-topbar')
  }

  document.getElementById('showActions').onclick = () => { showSection('Actions') }
  document.getElementById('showLogs').onclick = () => { showSection('Logs') }
}

function marshalDashboardStructureToHtml (json) {
  const nav = document.getElementById('navigation-links')

  for (const dashboard of json.dashboards) {
    const oldsection = document.querySelector('section[title="' + dashboard.title + '"]')

    if (oldsection != null) {
      oldsection.remove()
    }

    const section = document.createElement('section')
    section.title = dashboard.title.replaceAll(' ', '')

    const def = createFieldset('default', section)
    section.appendChild(def)

    document.getElementsByTagName('main')[0].appendChild(section)
    marshalContainerContents(dashboard, section, def, dashboard.title)

    const oldLi = nav.querySelector('li[title="' + dashboard.title + '"]')

    if (oldLi != null) {
      oldLi.remove()
    }

    const navigationA = document.createElement('a')
    navigationA.title = dashboard.title
    navigationA.innerText = dashboard.title
    navigationA.setAttribute('href', '#' + dashboard.title.replace(' ', ''))
    navigationA.onclick = () => {
      showSection(dashboard.title.replace(' ', ''))
    }

    const navigationLi = document.createElement('li')
    navigationLi.appendChild(navigationA)
    navigationLi.title = dashboard.title

    document.getElementById('navigation-links').appendChild(navigationLi)
  }

  const rootGroup = document.querySelector('#root-group')

  for (const btn of Object.values(window.actionButtons)) {
    if (btn.parentElement === null) {
      rootGroup.appendChild(btn)
    }
  }

  if (window.currentSection !== '') {
    showSection(window.currentSection)
  } else if (window.initialHash !== '' && document.body.getAttribute('initial-marshal-complete') === null) {
    showSection(window.initialHash.replace('#', ''))
  } else {
    if (rootGroup.querySelectorAll('action-button').length === 0 && json.dashboards.length > 0) {
      nav.querySelector('li[title="Actions"]').style.display = 'none'

      showSection(json.dashboards[0].title)
    } else {
      showSection('Actions')
    }
  }
}

function marshalLink (item, fieldset) {
  let btn = window.actionButtons[item.title]

  if (typeof btn === 'undefined') {
    btn = document.createElement('button')
    btn.innerText = 'Action not found: ' + item.title
    btn.classList.add('error')
  }

  fieldset.appendChild(btn)
}

function marshalMreOutput (dashboardComponent, fieldset) {
  const pre = document.createElement('pre')
  pre.classList.add('mre-output')
  pre.innerHTML = 'Waiting...'

  const executionStatus = {
    actionId: dashboardComponent.title
  }

  window.fetch(window.restBaseUrl + 'ExecutionStatus', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(executionStatus)
  }).then((res) => {
    if (res.ok) {
      return res.json()
    } else {
      pre.innerHTML = 'error'

      throw new Error(res.statusText)
    }
  }).then((json) => {
    updateMre(pre, json.logEntry)
  })

  const updateMre = (pre, json) => {
    pre.innerHTML = json.stdout
  }

  window.addEventListener('ExecutionFinished', (e) => {
    // The dashboard component "title" field is used for lots of things
    // and in this context for MreOutput it's just to refer an an actionId.
    //
    // So this is not a typo.
    if (e.payload.actionId === dashboardComponent.title) {
      updateMre(pre, e.payload)
    }
  })

  fieldset.appendChild(pre)
}

function marshalContainerContents (json, section, fieldset, parentDashboard) {
  for (const item of json.contents) {
    switch (item.type) {
      case 'fieldset':
        marshalFieldset(item, section, parentDashboard)
        break
      case 'directory':
        marshalDirectoryButton(item, fieldset)
        marshalDirectory(item, section)
        break
      case 'display':
        marshalDisplay(item, fieldset)
        break
      case 'stdout-most-recent-execution':
        marshalMreOutput(item, fieldset)
        break
      case 'link':
        marshalLink(item, fieldset)
        break
      default:
    }
  }
}

function createFieldset (title, parentDashboard) {
  const legend = document.createElement('legend')
  legend.innerText = title

  const fs = document.createElement('fieldset')
  fs.title = title
  fs.appendChild(legend)

  if (typeof parentDashboard === 'undefined') {
    fs.setAttribute('parent-dashboard', '')
  } else {
    fs.setAttribute('parent-dashboard', parentDashboard)
  }

  return fs
}

function marshalFieldset (item, section, parentDashboard) {
  const fs = createFieldset(item.title, parentDashboard)

  marshalContainerContents(item, section, fs)

  section.appendChild(fs)
}

function changeDirectory (selected) {
  if (selected === '') {
    selected = null
  }

  if (selected === null) {
    window.directoryNavigation = []
  } else if (selected === '..') {
    window.directoryNavigation.pop()

    if (window.directoryNavigation.length > 0) {
      selected = window.directoryNavigation[window.directoryNavigation.length - 1]
    } else {
      selected = null
    }
  } else {
    // If the selected item is already in the nav list, pop elements until we get
    // "back" to the existing nav item
    while (window.directoryNavigation.includes(selected)) {
      window.directoryNavigation.pop()
    }

    window.directoryNavigation.push(selected)
  }

  for (const fieldset of document.querySelectorAll('fieldset')) {
    if (selected === null) {
      if ((fieldset.id === 'root-group' || fieldset.getAttribute('parent-dashboard') !== '') && fieldset.children.length > 1) {
        fieldset.style.display = 'grid'
      } else {
        fieldset.style.display = 'none'
      }
    } else {
      if (fieldset.title === selected) {
        fieldset.style.display = 'grid'
      } else {
        fieldset.style.display = 'none'
      }
    }
  }

  const title = document.querySelector('h1')
  title.innerHTML = ''

  const rootLink = createDirectoryBreadcrumb(window.pageTitle, null)
  title.appendChild(rootLink)

  for (const dir of window.directoryNavigation) {
    const sep = document.createElement('span')
    sep.innerHTML = ' &raquo; '
    title.append(sep)

    if (dir === selected) {
      title.append(selected)
    } else {
      title.appendChild(createDirectoryBreadcrumb(dir))
    }
  }

  document.title = title.innerText

  if (selected === null) {
    window.history.pushState({ dir: null }, null, '#')
  } else {
    window.history.pushState({ dir: selected }, null, '#' + selected)
  }
}

function createDirectoryBreadcrumb (title, link) {
  const a = document.createElement('a')
  a.innerText = title
  a.title = title

  if (typeof link === 'undefined') {
    link = title
  }

  if (link === null) {
    a.href = '#'
  } else {
    a.href = '#' + link
  }

  a.onclick = () => {
    changeDirectory(link)
  }

  return a
}

function marshalDisplay (item, fieldset) {
  const display = document.createElement('div')
  display.innerHTML = item.title
  display.classList.add('display')

  fieldset.appendChild(display)
}

function marshalDirectoryButton (item, fieldset) {
  const directoryButton = document.createElement('button')
  directoryButton.innerHTML = '<span class = "icon">&#128193;</span> ' + item.title
  directoryButton.onclick = () => {
    changeDirectory(item.title)
  }

  fieldset.appendChild(directoryButton)
}

function marshalDirectory (item, section) {
  const fs = createFieldset(item.title)
  fs.style.display = 'none'

  const directoryBackButton = document.createElement('button')
  directoryBackButton.innerHTML = '&laquo;'
  directoryBackButton.title = 'Go back one directory'
  directoryBackButton.onclick = () => {
    changeDirectory('..')
  }

  fs.appendChild(directoryBackButton)

  marshalContainerContents(item, section, fs)

  section.appendChild(fs)
}

export function marshalLogsJsonToHtml (json) {
  for (const logEntry of json.logs) {
    const existing = window.logEntries[logEntry.executionTrackingId]

    if (existing !== undefined) {
      continue
    }

    window.logEntries[logEntry.executionTrackingId] = logEntry

    const tpl = document.getElementById('tplLogRow')
    const row = tpl.content.querySelector('tr').cloneNode(true)

    if (logEntry.stdout.length === 0) {
      logEntry.stdout = '(empty)'
    }

    if (logEntry.stderr.length === 0) {
      logEntry.stderr = '(empty)'
    }

    let logTableExitCode = logEntry.exitCode

    if (logEntry.exitCode === 0) {
      logTableExitCode = 'OK'
    }

    if (logEntry.timedOut) {
      logTableExitCode += ' (timed out)'
    }

    row.querySelector('.timestamp').innerText = logEntry.datetimeStarted
    row.querySelector('.content').innerText = logEntry.actionTitle
    row.querySelector('.icon').innerHTML = logEntry.actionIcon
    row.querySelector('pre.stdout').innerText = logEntry.stdout
    row.querySelector('pre.stderr').innerText = logEntry.stderr
    row.querySelector('.exit-code').innerText = logTableExitCode
    row.setAttribute('title', logEntry.actionTitle)

    row.querySelector('.content').onclick = () => {
      window.executionDialog.reset()
      window.executionDialog.show()
      window.executionDialog.renderExecutionResult({
        logEntry: window.logEntries[logEntry.executionTrackingId]
      })
    }

    for (const tag of logEntry.tags) {
      const domTag = document.createElement('span')
      domTag.classList.add('tag')
      domTag.innerText = tag

      row.querySelector('.tags').append(domTag)
    }

    document.querySelector('#logTableBody').prepend(row)
  }
}

window.addEventListener('popstate', (e) => {
  e.preventDefault()
  if (e.state != null && typeof e.state.dir !== 'undefined') {
    changeDirectory(e.state.dir)
  }
})
