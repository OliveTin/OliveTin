import { describe, it, before, after } from 'mocha'
import { assert } from 'chai'
import {
  getRootAndWait,
  getActionButtons,
  takeScreenshotOnFailure,
} from '../../lib/elements.js'

describe('config: inlineActions', function () {
  before(async function () {
    await runner.start('inlineActions')
  })

  after(async () => {
    await runner.stop()
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver);
  });

  it('Inline dashboard actions are rendered as clickable buttons', async function () {
    await getRootAndWait()

    const buttons = await getActionButtons()
    assert.isArray(buttons, 'Action buttons should be an array')
    assert.isAtLeast(buttons.length, 1, 'There should be at least one action button')

    const texts = await Promise.all(buttons.map(b => b.getText()))
    const combinedText = texts.join(' ')

    assert.include(
      combinedText,
      'Inline Dashboard Action',
      'Inline dashboard action should be rendered as a button'
    )
  })
})