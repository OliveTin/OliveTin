import * as process from 'node:process'
import waitOn from 'wait-on'
import { spawn } from 'node:child_process'

export default function getRunner () {
  const type = process.env.OLIVETIN_TEST_RUNNER

  console.log('OLIVETIN_TEST_RUNNER env value is: ', type)

  switch (type) {
    case 'local':
      return new OliveTinTestRunnerStartLocalProcess()
    case 'vm':
      return new OliveTinTestRunnerVm()
    case 'container':
      return new OliveTinTestRunnerEnv()
    default:
      return new OliveTinTestRunnerStartLocalProcess()
  }
}

class OliveTinTestRunner {
  BASE_URL = 'http://nohost:1337/';

  baseUrl() {
    return this.BASE_URL
  }
}

class OliveTinTestRunnerStartLocalProcess extends OliveTinTestRunner {
  async start (cfg) {
    this.ot = spawn('./../OliveTin', ['-configdir', 'configs/' + cfg + '/'])

    const logStdout = process.env.OLIVETIN_TEST_RUNNER_LOG_STDOUT === '1'

    if (logStdout) {
      this.ot.stdout.on('data', (data) => {
        console.log(`stdout: ${data}`)
      })

      this.ot.stderr.on('data', (data) => {
        console.error(`stderr: ${data}`)
      })
    }

    this.ot.on('close', (code) => {
      if (code != null) {
        console.log(`child process exited with code ${code}`)
      }
    })

    /*
      this.server = await startSomeServer({port: process.env.TEST_PORT});
      console.log(`server running on port ${this.server.port}`);
      */

    this.BASE_URL = 'http://localhost:1337/'

    await waitOn({
      resources: [this.BASE_URL]
    })
  }

  async stop () {
    await this.ot.kill()
  }
}

class OliveTinTestRunnerEnv extends OliveTinTestRunner {
  constructor () {
    super()

    const IP = process.env.IP
    const PORT = process.env.PORT

    this.BASE_URL = 'http://' + IP + ':' + PORT + '/'

    console.log('Runner ENV endpoint: ' + this.BASE_URL)
  }

  async start () {
    await waitOn({
      resources: [this.BASE_URL]
    })
  }

  async stop () {

  }
}

class OliveTinTestRunnerVm extends OliveTinTestRunnerEnv {
  constructor() {
    super()
  }

  async start (cfg) {
    console.log("vagrant changing config")
    spawn('vagrant', ['ssh', '-c', '"ln -sf /etc/OliveTin/ /opt/OliveTin-configs/' + cfg + '/config.yaml"'])
    spawn('vagrant', ['ssh', '-c', '"systemctl restart OliveTin"'])

    return null
  }
}
