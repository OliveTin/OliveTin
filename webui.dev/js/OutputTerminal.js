import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'

export class OutputTerminal {
  init () {
    this.termianl = new Terminal({
      convertEol: true
    })

    const fitAddon = new FitAddon()
    this.terminal.loadAddon(fitAddon)
    this.terminal.fit = fitAddon
  }

  write (out, then) {
    this.terminal.write(out, then)
  }

  fit () {
    this.terminal.fit.fit()
  }

  open (el) {
    this.terminal.open(el)
  }

  reset () {
    this.terminal.reset()
  }
}
