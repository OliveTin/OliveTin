import process from 'node:process'
import waitOn from 'wait-on'
import { spawn } from 'node:child_process'

let ot = null

export default function getRunner () {
  const type = process.env.OLIVETIN_TEST_RUNNER

  console.log('OLIVETIN_TEST_RUNNER env value is: ', type)

  switch (type) {
    case 'local':
      return new OliveTinTestRunnerLocalProcess()
    case 'vm':
      return null
    case 'container':
      return null
    default:
      return new OliveTinTestRunnerLocalProcess()
  }
}

class OliveTinTestRunnerLocalProcess {
  async start (cfg) {
    ot = spawn('./../OliveTin', ['-configdir', 'configs/' + cfg + '/'])

    const logStdout = process.env.OLIVETIN_TEST_RUNNER_LOG_STDOUT === '1'

    if (logStdout) {
      ot.stdout.on('data', (data) => {
        console.log(`stdout: ${data}`)
      })

      ot.stderr.on('data', (data) => {
        console.error(`stderr: ${data}`)
      })
    }

    ot.on('close', (code) => {
      if (code != null) {
        console.log(`child process exited with code ${code}`)
      }
    })

    /*
      this.server = await startSomeServer({port: process.env.TEST_PORT});
      console.log(`server running on port ${this.server.port}`);
      */

    await waitOn({
      'resources': ['http://localhost:1337/']
    })

    return ot
  }

  async stop () {
    await ot.kill()
  }
}
