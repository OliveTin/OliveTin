import { expect } from 'chai'

describe('config: trustedHeader', function () {
  before(async function () {
    await runner.start('trustedHeader')
  })

  after(async () => {
    await runner.stop()
  })

  it('req with X-User', async () => {
    const req = await fetch(runner.baseUrl() + '/api/WhoAmI', {
      headers: {
        "X-User": "fred",
      }
    })

    if (!req.ok) {
      console.log(req)
    }

    expect(req.ok, 'WhoAmI Request is ' + req.status).to.be.true

    const json = await req.json()

    expect(json).to.not.be.null
    expect(json).to.have.own.property('authenticatedUser')

    expect(json['authenticatedUser']).to.be.equal('fred')
  })
})
