import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'

import { By } from 'selenium-webdriver'

let metrics = [
  {'name': 'olivetin_actions_requested_count', 'type': 'counter', 'desc': 'The actions requested count'},
  {'name': 'olivetin_config_action_count', 'type': 'gauge', 'desc': 'The number of actions in the config file'},
  {'name': 'olivetin_config_reloaded_count', 'type': 'counter', 'desc': 'The number of times the config has been reloaded'},
  {'name': 'olivetin_sv_count', 'type': 'gauge', 'desc': 'The number entries in the sv map'},
]

describe('config: prometheus', function () {
  before(async function () {
    await runner.start('prometheus')
  })

  after(async () => {
    await runner.stop()
  })

  it('Metrics are available with correct types', async () => {
    webdriver.get(runner.metricsUrl())
    const prometheusOutput = await webdriver.findElement(By.tagName('pre')).getText()

    expect(prometheusOutput).to.not.be.null
    metrics.forEach(({name, type, desc}) => {
      const metaLines = `# HELP ${name} ${desc}\n`
        + `# TYPE ${name} ${type}\n`
      expect(prometheusOutput).to.match(new RegExp(metaLines))
    })
  })
})
