// Issue: https://github.com/OliveTin/OliveTin/issues/616
import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import { 
  getRootAndWait, 
  getActionButtons,
  closeExecutionDialog,
  takeScreenshotOnFailure,
  getExecutionDialogOutput,
} from '../lib/elements.js'

describe('config: entities', function () {
  before(async function () {
    await runner.start('entityFilesWithLongIntsUseStandardForm')
  })

  after(async () => {
    await runner.stop()
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver);
  });

  it('Entity buttons are rendered', async function() {
    await getRootAndWait()

    const buttons = await getActionButtons()

    expect(buttons).to.not.be.null
    expect(buttons).to.have.length(5)

    const buttonInt10 = await buttons[2]   
    expect(await buttonInt10.getAttribute('title')).to.be.equal('Test me INT with 10 numbers')
    await buttonInt10.click()
    expect(await getExecutionDialogOutput()).to.be.equal('1234567890\n', 'Expected output to be an int')

    await closeExecutionDialog()

    const buttonFloat10 = await buttons[0]
    expect(await buttonFloat10.getAttribute('title')).to.be.equal('Test me FLOAT with 10 numbers')
    await buttonFloat10.click()
    expect(await getExecutionDialogOutput()).to.be.equal('1.234568\n', 'Expected output to be a float')

  });
});
