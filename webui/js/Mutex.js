export class Mutex {
  constructor() {
    this._locked = false;
    this._waiting = [];
  }

  lock() {
    const unlock = () => {
      const next = this._waiting.shift();
      if (next) {
        next(unlock);
      } else {
        this._locked = false;
      }
    };

    if (this._locked) {
      return new Promise(resolve => this._waiting.push(resolve)).then(() => unlock);
    } else {
      this._locked = true;
      return Promise.resolve(unlock);
    }
  }
}
