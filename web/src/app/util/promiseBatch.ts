import { BATCH_DELAY } from '../config'

function _finally<T>(p: Promise<T>, fn: () => void): Promise<T> {
  if (p.finally) return p.finally(fn)

  // fallback to manual implementation
  return p.then(
    (val) => {
      const v = (): T => val
      return Promise.resolve().then(fn).then(v)
    },
    (err) => {
      const e = (): Promise<T> => Promise.reject(err)
      return Promise.resolve().then(fn).then(e)
    },
  )
}

// BatchPromise allows batching promises together so they resolve at the same time.
// It differs from Promise.all and Promise.allSettled in that you can add
// additional promises after creation.
class BatchPromise {
  private _p: Promise<void>
  private _resolveP!: () => void
  private _unresolvedCount = 1 // start with +1 for the timer

  constructor() {
    this._p = new Promise((resolve) => {
      this._resolveP = resolve
    })
  }

  resolveOne = (): Promise<void> => {
    this._unresolvedCount--
    if (this._unresolvedCount === 0) this._resolveP()
    return this._p
  }

  addPromise<T>(p: Promise<T>): Promise<T> {
    this._unresolvedCount++
    return _finally(p, this.resolveOne)
  }
}

let currentBatch: BatchPromise | null = null
let t: NodeJS.Timeout
function batchReset(): void {
  currentBatch?.resolveOne()
  currentBatch = null
}

// promiseBatch will return a new Promise that will resolve after all Promises
// that were provided during the debounce (ms) delay.
export default function promiseBatch<T>(
  p: Promise<T>,
  debounce: number = BATCH_DELAY,
): Promise<T> {
  if (!currentBatch) {
    currentBatch = new BatchPromise()
  }

  clearTimeout(t)
  t = setTimeout(batchReset, debounce)

  return currentBatch.addPromise(p)
}
