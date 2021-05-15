'use strict'

import { marshalActionButtonsJsonToHtml } from './js/marshaller.js'

/**
 * Design choice; define this as a "global function" (on window) so that it can 
 * easily be used anywhere without needing to import it, as it's a pretty 
 * fundemental thing.
 */
window.showBigError = (type, friendlyType, message) => {
	console.error("Error " + type + ": ", err)

	var err = document.createElement("div")
	err.classList.add("error")
	err.innerHTML = "<h1>Error " + friendlyType + "</h1><p>" + message + "</p><p><a href = 'http://github.com/jamesread/OliveTin' target = 'blank'/>OliveTin Documentation</a></p>";

  document.getElementById('rootGroup').appendChild(err)
}

function onInitialLoad(res) {
  window.restBaseUrl = res.Rest;

  window.fetch(window.restBaseUrl + "GetButtons", {
    cors: 'cors',
    // No fetch options
  }).then(res => {
    return res.json()
  }).then(res => {
    marshalActionButtonsJsonToHtml(res)
  }).catch(err => {
    showBigError("fetch-initial-buttons", "getting initial buttons", err, "blat")
  });
}

window.fetch('webUiSettings.json').then(res => {
  return res.json()
}).then(res => {
  onInitialLoad(res)
}).catch(err => {
  showBigError("fetch-webui-settings", "getting webui settings", err);
})
