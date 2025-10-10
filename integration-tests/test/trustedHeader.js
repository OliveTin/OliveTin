import { expect } from 'chai'
import { 
  getRootAndWait,
  takeScreenshotOnFailure,
} from '../lib/elements.js'

describe('config: trustedHeader', function () {
  before(async function () {
    await runner.start('trustedHeader')
  })

  after(async () => {
    await runner.stop()
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver);
  });

  it.skip('req with X-User', async () => {
    await getRootAndWait()

    // Use the Connect RPC client format
    const req = await fetch(runner.baseUrl() + '/api/Init', {
      method: 'POST',
      headers: {
        "X-User": "fred",
        "Content-Type": "application/json",
      },
      body: JSON.stringify({}),
    })

    console.log(`Final URL: ${req.url}, Status: ${req.status}`)

    if (!req.ok) {
      console.log('Request failed:', req.status, req.statusText)
      const text = await req.text()
      console.log('Response body:', text)
    }

    expect(req.ok, 'Init Request is ' + req.status).to.be.true

    const json = await req.json()

    expect(json).to.not.be.null
    expect(json).to.have.own.property('authenticatedUser')

    expect(json['authenticatedUser']).to.be.equal('fred')
  })
})
