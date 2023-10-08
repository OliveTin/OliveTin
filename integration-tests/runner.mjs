import process from 'node:process'
import {spawn} from 'node:child_process'

let ot = null;

export default function getRunner() {
  const type = process.env.OLIVETIN_TEST_RUNNER

  console.log("TEST RUNNER IS: ", type)

  switch (type) {
    case 'local':
      return new OliveTinTestRunnerLocalProcess();
    case 'vm':
      return null;
    case 'container':
      return null;
    default:
      console.warn('Using default test runner')

      return new OliveTinTestRunnerLocalProcess();
  }
}

class OliveTinTestRunnerLocalProcess {
  start(cfg) {
    ot = spawn("./../OliveTin", ['-configdir', 'configs/' + cfg + '/'])

    const logStdout = process.env.OLIVETIN_TEST_RUNNER_LOG_STDOUT == 1

    if (logStdout) {
      ot.stdout.on('data', (data) => {
          console.log(`stdout: ${data}`);
      });

      ot.stderr.on('data', (data) => {
          console.error(`stderr: ${data}`);
      });
    }


    ot.on('close', (code) => {
        console.log(`child process exited with code ${code}`);
    });

    /*
      this.server = await startSomeServer({port: process.env.TEST_PORT});
      console.log(`server running on port ${this.server.port}`);
      */
    return ot
  }

  stop() {
    ot.kill();
  }
}
