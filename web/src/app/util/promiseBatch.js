const BATCH_DELAY = 10

let current, timeout
function Batch(initActive = 0) {
  this._active = initActive
  this.promise = new Promise(resolve => {
    this._resolve = resolve
  })
  this.start = () => {
    this._active++
  }
  this.done = () => {
    this._active--
    if (this._active === 0) this._resolve()
  }
}

function start() {
  if (!current) current = new Batch(1)
  current.start()
  clearTimeout(timeout)
  const done = current.done
  timeout = setTimeout(() => {
    current = null
    done()
  }, BATCH_DELAY)
  return [current.promise, current.done]
}

// promiseBatch will wrap a promise so that it resolves at the same time as any
// others within the delay.
export default function promiseBatch(p) {
  const [wait, done] = start()
  return p.finally(done).finally(() => wait)
}
