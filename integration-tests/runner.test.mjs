import test from 'node:test'
import assert from 'node:assert/strict'
import getRunner from './runner.mjs'

test('OliveTinTestRunnerVm.start advances pageGeneration', async (t) => {
  const originalRunner = process.env.OLIVETIN_TEST_RUNNER
  const originalIp = process.env.IP
  const originalPort = process.env.PORT

  t.after(() => {
    if (originalRunner === undefined) {
      delete process.env.OLIVETIN_TEST_RUNNER
    } else {
      process.env.OLIVETIN_TEST_RUNNER = originalRunner
    }
    if (originalIp === undefined) {
      delete process.env.IP
    } else {
      process.env.IP = originalIp
    }
    if (originalPort === undefined) {
      delete process.env.PORT
    } else {
      process.env.PORT = originalPort
    }
  })

  process.env.OLIVETIN_TEST_RUNNER = 'vm'
  process.env.IP = '127.0.0.1'
  process.env.PORT = '1337'

  const runner = getRunner()
  const before = runner.pageGeneration

  await runner.start('pageGenerationFirst')

  assert.equal(runner.pageGeneration, before + 1)
})
