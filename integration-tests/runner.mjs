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

  metricsUrl() {
    return new URL('metrics', this.baseUrl());
  }
}

class OliveTinTestRunnerStartLocalProcess extends OliveTinTestRunner {
  async start (cfg) {
    let stdout = ""
    let stderr = ""

    console.log("      OliveTin starting local process...")

    this.ot = spawn('./../service/OliveTin', ['-configdir', 'configs/' + cfg + '/'])

    let logStdout = false

    if (process.env.CI === 'true') {
      logStdout = true;
    } else {
      logStdout = process.env.OLIVETIN_TEST_RUNNER_LOG_STDOUT === '1'
    }

    this.ot.stdout.on('data', (data) => {
      stdout += data

      if (logStdout) {
        console.log(`stdout: ${data}`)
      }
    })

    this.ot.stderr.on('data', (data) => {
      stderr += data

      if (logStdout) {
        console.log(`stderr: ${data}`)
      }
    })

    this.ot.on('close', (code) => {
      if (code != null) {
        console.log(`OliveTin local process exited with code ${code}`)
        console.log(stdout)
        console.log(stderr)
        console.log(this.ot.exitCode)
      }
    })

    if (this.ot.exitCode == null) {
      this.BASE_URL = 'http://localhost:1337/'

      console.log("      OliveTin waiting for local process to start...")

      await waitOn({
        resources: [this.BASE_URL]
      })

      console.log("      OliveTin local process started and webUI accessible")
    } else {
      console.log("      OliveTin local process start FAILED!")
      console.log(stdout)
      console.log(stderr)
      console.log(this.ot.exitCode)
    }
  }

  async stop () {
    if ((await this.ot.exitCode) != null) {
      console.log("      OliveTin local process tried stop(), but it already exited with code", this.ot.exitCode)
    } else {
      await this.ot.kill()
      console.log("      OliveTin local process killed")
    }

    if (process.env.CI === 'true') {
      // GitHub runners seem to need a bit more time to clean up
      await new Promise((res) => setTimeout(res, 3000))
    } else {
      await new Promise((res) => setTimeout(res, 100))
    }
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
