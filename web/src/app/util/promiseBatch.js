import { BATCH_DELAY } from '../config'

function _finally(p, fn) {
  if (p.finally) return p.finally(fn)

  // fallback to manual implementation
  return p.then(
    (val) => {
      const v = () => val
      return Promise.resolve().then(fn).then(v)
    },
    (err) => {
      const e = () => Promise.reject(err)
      return Promise.resolve().then(fn).then(e)
    },
  )
}

// BatchPromise allows batching promises together so they resolve at the same time.
// It differs from Promise.all and Promise.allSettled in that you can add
// additional promises after creation.
class BatchPromise {
  _p = new Promise((resolve) => {
    this._resolveP = resolve
  })

  _unresolvedCount = 1 // start with +1 for the timer

  resolveOne = () => {
    this._unresolvedCount--
    if (this._unresolvedCount === 0) this._resolveP()
    return this._p
  }

  addPromise = (p) => {
    this._unresolvedCount++
    return _finally(p, this.resolveOne)
  }
}

let currentBatch, t
function batchReset() {
  currentBatch.resolveOne()
  currentBatch = null
}

// promiseBatch will return a new Promise that will resolve after all Promises
// that were provided during the debounce (ms) delay.
export default function promiseBatch(p, debounce = BATCH_DELAY) {
  if (!currentBatch) {
    currentBatch = new BatchPromise()
  }

  clearTimeout(t)
  t = setTimeout(batchReset, debounce)

  return currentBatch.addPromise(p)
}
