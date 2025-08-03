import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { Mutex } from './Mutex.js'

/** 
 * xterm.js based terminal output for the execution dialog.
 *
 * the xterm.js methods for write(), reset() and clear() appear to be async,
 * but they do not return a Promise and instead use a callback. When calling
 * these methods in quick succession, the output can get garbled due to race 
 * conditions. 
 *
 * To avoid this, this class uses Mutex around those methods to ensure that 
 * only one write OR reset is executed at a time, is completed, and the calls
 * occour in sequential order.
 */
export class OutputTerminal {
  constructor () {
    this.writeMutex = new Mutex()
    this.terminal = new Terminal({
      convertEol: true
    })

    const fitAddon = new FitAddon()
    this.terminal.loadAddon(fitAddon)
    this.terminal.fit = fitAddon
  }

  async write (out, then) {
    const unlock = await this.writeMutex.lock()

    try {
      await new Promise(resolve => {
        this.terminal.write(out, () => {
          resolve()
        })
      })
    } finally {
      unlock()

      if (then != null && then !== undefined) {
        then()
      }
    }
  }

  async reset () {
    const unlock = await this.writeMutex.lock()

    try {
      await new Promise(resolve => {
        this.terminal.clear()
        this.terminal.reset()
        resolve()
      })
    } finally {
      unlock()
    }
  }

  fit () {
    this.terminal.fit.fit()
  }

  open (el) {
    this.terminal.open(el)
  }

  resize (cols, rows) {
    this.terminal.resize(cols, rows)
  }
}
