import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import { By, until } from 'selenium-webdriver'
import { getActionButtons, getRootAndWait } from '../lib/elements.js'

describe('config: sleep', function () {
  before(async function () {
    await runner.start('sleep')
  })

  after(async () => {
    await runner.stop()
  })

  it('Sleep action kill', async function() {
    await getRootAndWait()

    const buttons = await getActionButtons(webdriver)

  })
})
