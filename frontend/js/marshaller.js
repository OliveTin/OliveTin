function checkAndTriggerActionFromQueryParam () {
  const params = getQueryParams()

  const action = params.get('action')
  if (action && window.actionButtons) {
    // Look for an action button with matching title
    const actionButton = window.actionButtons[action]

    if (actionButton) {
      // Only trigger actions that have arguments
      const jsonButton = window.actionButtonsJson[action]
      if (jsonButton && jsonButton.arguments && jsonButton.arguments.length > 0) {
        // Trigger the action button click
        setTimeout(() => {
          actionButton.btn.click()
        }, 500) // Small delay to ensure UI is fully loaded
        return true
      }
    }
  }
  return false
}

function createElement (tag, attributes) {
  const el = document.createElement(tag)

  if (attributes !== null) {
    if (attributes.classNames !== undefined) {
      el.classList.add(...attributes.classNames)
    }

    if (attributes.innerText !== undefined) {
      el.innerText = attributes.innerText
    }
  }

  return el
}

function createTag (val) {
  const domTag = createElement('span', {
    innerText: val,
    classNames: ['tag']
  })

  return domTag
}

function createAnnotation (key, val) {
  const domAnnotation = createElement('span', {
    classNames: ['annotation']
  })

  domAnnotation.appendChild(createElement('span', {
    innerText: key,
    classNames: ['annotation-key']
  }))

  domAnnotation.appendChild(createElement('span', {
    innerText: val,
    classNames: ['annotation-value']
  }))

  return domAnnotation
}

/**
 * This is a weird function that just sets some globals.
 */
export function initMarshaller () {
  window.logEntries = new Map()


  window.addEventListener('EventExecutionStarted', onExecutionStarted)
  window.addEventListener('EventExecutionFinished', onExecutionFinished)
  window.addEventListener('EventOutputChunk', onOutputChunk)
}

function setUsername (username, provider) {
  document.getElementById('username').innerText = username
  document.getElementById('username').setAttribute('title', provider)

  if (window.settings.AuthLocalLogin || window.settings.AuthOAuth2Providers !== null) {
    if (username === 'guest') {
      document.getElementById('link-login').hidden = false
      document.getElementById('link-logout').hidden = true
    } else {
      document.getElementById('link-login').hidden = true

      if (provider === 'local' || provider === 'oauth2') {
        document.getElementById('link-logout').hidden = false
      }
    }
  }
}

export function marshalDashboardComponentsJsonToHtml (json) {
  if (json == null) { // eg: HTTP 403
    setUsername('guest', 'system')

    if (window.settings.AuthLoginUrl !== '') {
      window.location = window.settings.AuthLoginUrl
    } else {
      showSection('/login')
    }
  } else {
    setUsername(json.authenticatedUser, json.authenticatedUserProvider)

    marshalDashboardStructureToHtml(json)
  }

  document.body.setAttribute('initial-marshal-complete', 'true')
}

function onOutputChunk (evt) {
  const chunk = evt.payload

  return;
  if (chunk.executionTrackingId === window.executionDialog.executionTrackingId) {
    window.terminal.write(chunk.output)
  }
}

function onExecutionStarted (evt) {
  const logEntry = evt.payload.logEntry

  // marshalLogsJsonToHtml({
  //   logs: [logEntry]
  // })
}

function onExecutionFinished (evt) {
  const logEntry = evt.payload.logEntry

  window.logEntries.set(logEntry.executionTrackingId, logEntry)

  return;

  const executionButton = document.querySelector('execution-button#execution-' + logEntry.executionTrackingId)
  let feedbackButton = actionButton

  switch (actionButton.popupOnStart) {
    case 'execution-button':
      if (executionButton != null) {
        feedbackButton = executionButton
      }

      break
    case 'execution-dialog-output-html':
    case 'execution-dialog-stdout-only':
    case 'execution-dialog':
      // We don't need to fetch the logEntry for the dialog because we already
      // have it, so we open the dialog and it will get updated below.

      window.executionDialog.show()
      window.executionDialog.executionTrackingId = logEntry.uuid

      break
  }

  feedbackButton.onExecutionFinished(logEntry)

  // marshalLogsJsonToHtml({
  //   logs: [logEntry]
  // })

  // If the current execution dialog is open, update that too
  if (window.executionDialog.dlg.open && window.executionDialog.executionUuid === logEntry.uuid) {
    window.executionDialog.renderExecutionResult({
      logEntry: logEntry,
      type: actionButton.popupOnStart
    })
  }
}

function convertPathToBreadcrumb (path) {
  const parts = path.split('/')

  const result = []

  for (let i = 0; i < parts.length; i++) {
    if (parts[i] === '') {
      continue
    }

    result.push(parts.slice(0, i + 1).join('/'))
  }

  return result
}

function showExecutionResult (pathName) {
  const executionTrackingId = pathName.split('/')[2]
  window.executionDialog.fetchExecutionResult(executionTrackingId)
  window.executionDialog.show()
}

function setSectionNavigationVisible (visible) {
  const nav = document.querySelector('nav')
  const btn = document.getElementById('sidebar-toggler-button')

  nav.removeAttribute('hidden')

  if (document.body.classList.contains('has-sidebar')) {
    if (visible) {
      btn.setAttribute('aria-pressed', false)
      btn.setAttribute('aria-label', 'Open sidebar navigation')
      btn.innerHTML = '&laquo;'

      nav.classList.add('shown')
    } else {
      btn.setAttribute('aria-pressed', true)
      btn.setAttribute('aria-label', 'Close sidebar navigation')
      btn.innerHTML = '&#9776;'

      nav.classList.remove('shown')
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
}

function marshalSingleDashboard (dashboard, nav) {
  const oldsection = document.querySelector('section[title="' + getSystemTitle(dashboard.title) + '"]')

  if (oldsection != null) {
    oldsection.remove()
  }

  const section = document.createElement('section')
  section.setAttribute('system-title', getSystemTitle(dashboard.title))
  section.title = section.getAttribute('system-title')

  const def = createFieldset('default', section)
  section.appendChild(def)

  document.getElementsByTagName('main')[0].appendChild(section)
  marshalContainerContents(dashboard, section, def, dashboard.title)

  const oldLi = nav.querySelector('li[title="' + dashboard.title + '"]')

  if (oldLi != null) {
    oldLi.remove()
  }

  const systemTitleUrl = '/' + getSystemTitle(dashboard.title)

  window.navbar.createLink(dashboard.title, systemTitleUrl, false)

}

function marshalDashboardStructureToHtml (json) {
  return
  const nav = document.getElementById('navigation-links')

  for (const dashboard of json.dashboards) {
    marshalSingleDashboard(dashboard, nav)
  }

  const rootGroup = document.querySelector('#root-group')

  for (const btn of Object.values(window.actionButtons)) {
    if (btn.parentElement === null) {
      rootGroup.appendChild(btn)
    }
  }

  const shouldHideActions = rootGroup.querySelectorAll('action-button').length === 0 && json.dashboards.length > 0

  if (shouldHideActions) {
    nav.querySelector('li[title="Actions"]').style.display = 'none'
  }

  if (window.currentPath !== '') {
    showSection(window.currentPath)
  } else if (window.location.pathname !== '/' && document.body.getAttribute('initial-marshal-complete') === null) {
    showSection(window.location.pathname)
  } else {
    if (shouldHideActions) {
      showSection('/' + getSystemTitle(json.dashboards[0].title))
    } else {
      showSection('/')
    }
  }
}

function marshalLink (item, fieldset) {
  return
  let btn = window.actionButtons[item.title]

  if (typeof btn === 'undefined') {
    btn = document.createElement('button')
    btn.innerText = 'Action not found: ' + item.title
    btn.classList.add('error')
  }

  if (item.cssClass !== '') {
    btn.classList.add(item.cssClass)
  }

  fieldset.appendChild(btn)
}

function marshalMreOutput (dashboardComponent, fieldset) {
  const pre = document.createElement('pre')
  pre.classList.add('mre-output')
  pre.innerHTML = 'Waiting...'

  const args = {
    actionId: dashboardComponent.title
  }

  try { 
    const status = window.client.executionStatus(args)

    updateMre(pre, status.logEntry)
  } catch (err) {
    pre.innerHTML = 'error'

      throw new Error(res.statusText)
  }

  const updateMre = (pre, json) => {
    pre.innerHTML = json.output
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
      case 'directory': {
        const directoryPath = marshalDirectory(item, section)
        marshalDirectoryButton(item, fieldset, directoryPath)
      }
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

  marshalContainerContents(item, section, fs, parentDashboard)

  section.appendChild(fs)
}

function createNavigationBreadcrumbDisplay (path) {
  const a = document.createElement('a')
  a.href = 'javascript:void(0)'

  if (path.view === null) {
    a.title = path.section
    a.innerText = path.section
  } else {
    a.innerText = path.view
    a.title = path.view
  }

  a.onclick = () => {
    showSectionView(path.view)
  }

  return a
}

function marshalDisplay (item, fieldset) {
  const display = document.createElement('div')
  display.innerHTML = item.title
  display.classList.add('display')

  if (item.cssClass !== '') {
    display.classList.add(item.cssClass)
  }

  fieldset.appendChild(display)
}

function marshalDirectoryButton (item, fieldset, path) {
  const directoryButton = document.createElement('button')
  directoryButton.innerHTML = '<span class = "icon">' + item.icon + '</span> ' + item.title
  directoryButton.onclick = () => {
    showSection(path)
  }

  fieldset.appendChild(directoryButton)
}

function marshalDirectory (item, section) {
  const fs = createFieldset(item.title)
  fs.style.display = 'none'

  const directoryBackButton = document.createElement('button')
  directoryBackButton.innerHTML = window.settings.DefaultIconForBack
  directoryBackButton.title = 'Go back one directory'
  directoryBackButton.onclick = () => {
    showSection('/' + section.title)
  }

  fs.appendChild(directoryBackButton)

  marshalContainerContents(item, section, fs)

  section.appendChild(fs)

  const path = '/' + section.title + '/' + getSystemTitle(item.title)

  return path
}

export function refreshServerConnectionLabel () {
  if (window.restAvailable) {
    document.querySelector('#serverConnectionRest').classList.remove('error')
  } else {
    document.querySelector('#serverConnectionRest').classList.add('error')
  }

  if (window.websocketAvailable) {
    document.querySelector('#serverConnectionWebSocket').classList.remove('error')
    document.querySelector('#serverConnectionWebSocket').innerText = 'WebSocket'
  } else {
    document.querySelector('#serverConnectionWebSocket').classList.add('error')
    document.querySelector('#serverConnectionWebSocket').innerText = 'WebSocket Error'
  }
}
