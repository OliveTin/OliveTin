'use strict'

import { loadContents } from './js/loader.js'

window.fetch('http://mindstorm4:1339/GetButtons').then(res => {
  return res.json()
}).then(res => {
  loadContents(res)
})
