import { useEffect, useMemo, useState } from 'react'
import { pathPrefix } from '../env'
import methods, { WorkerMethod, WorkerParam } from './methods'

type RecvMessage<M extends WorkerMethod> = {
  data: ReturnType<M>
}

type NextRun<M extends WorkerMethod> = {
  arg: WorkerParam<M>
}

type ChangeCallback<M extends WorkerMethod> = (result: ReturnType<M>) => void

type Post<M extends WorkerMethod> = {
  method: string
  arg: WorkerParam<M>
}

// StubWorker does work after a setTimeout, but in the main thread.
class StubWorker<M extends WorkerMethod> {
  constructor(method: M) {
    this.method = method
  }

  private method: M
  private _timeout: ReturnType<typeof setTimeout> | undefined
  onmessage: (e: RecvMessage<M>) => void = (): void => {}

  postMessage = (data: Post<M>): void => {
    this._timeout = setTimeout(() => {
      this.onmessage({
        data: this.method(data.arg) as ReturnType<M>,
      })
    })
  }

  terminate = (): void => {
    if (!this._timeout) return
    clearTimeout(this._timeout)
  }
}

class Runner<M extends WorkerMethod> {
  constructor(method: M, onChange: ChangeCallback<M>) {
    this.method = method
    this.onChange = onChange
  }

  private method: M
  private worker: Worker | StubWorker<M> | null = null
  private next: NextRun<M> | null = null
  private onChange: ChangeCallback<M>
  private isBusy = false

  private _initWorker = (): Worker | StubWorker<M> => {
    const w = window.Worker
      ? new Worker(`${pathPrefix}/static/worker.js`)
      : new StubWorker(this.method)

    w.onmessage = (e: RecvMessage<M>) => {
      this.isBusy = false
      this.onChange(e.data)
      this._send()
    }

    return w
  }

  private _send = (): void => {
    if (!this.next) return
    if (!this.worker) {
      this.worker = this._initWorker()
    }
    if (this.isBusy) return
    this.worker.postMessage({ method: this.method.name, arg: this.next.arg })
    this.isBusy = true
    this.next = null
  }

  run = (arg: WorkerParam<M>): void => {
    this.next = { arg }
    this._send()
  }

  shutdown = (): void => {
    if (!this.worker) return
    this.worker.terminate()
    this.worker = null
  }
}

export function useWorker<M extends WorkerMethod>(
  method: M,
  arg: WorkerParam<M>,
  def: ReturnType<M>,
): ReturnType<M> {
  if (!(method.name in methods)) {
    throw new Error(`method must be a valid method from app/worker/methods.ts`)
  }

  const [result, setResult] = useState(def)
  const [worker, setWorker] = useState<Runner<M> | null>(null)

  useEffect(() => {
    const w = new Runner<M>(method, setResult)
    setWorker(w)
    return w.shutdown
  }, [])

  useMemo(() => {
    if (!worker) return
    worker.run(arg)
  }, [worker, arg])

  return result
}
