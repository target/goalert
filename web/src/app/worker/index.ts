import { useEffect, useMemo, useState } from 'react'
import { pathPrefix } from '../env'
import methods, {
  WorkerMethod,
  WorkerMethodName,
  WorkerReturnType,
  WorkerParam,
  WorkerResult,
} from './methods'

type RecvMessage<N extends WorkerMethodName> = {
  data: WorkerResult<N>
}

type NextRun<N extends WorkerMethodName> = {
  arg: WorkerParam<N>
}

type ChangeCallback<N extends WorkerMethodName> = (
  result: WorkerResult<N>,
) => void

type Post<N extends WorkerMethodName> = {
  method: N
  arg: WorkerParam<N>
}

type OnMessage<N extends WorkerMethodName> = (e: RecvMessage<N>) => void

// StubWorker does work after a setTimeout, but in the main thread.
class StubWorker<N extends WorkerMethodName> {
  constructor(methodName: N) {
    this.method = methods[methodName]
  }

  private method: WorkerMethod<N>
  private _timeout: ReturnType<typeof setTimeout> | undefined
  onmessage: OnMessage<N> = (): void => {}

  postMessage = (data: Post<N>): void => {
    this._timeout = setTimeout(() => {
      this.onmessage({
        // typescript calculates the incorrect type for the method argument
        /* eslint-disable @typescript-eslint/no-explicit-any */
        data: this.method(data.arg as any) as WorkerResult<N>,
      })
    })
  }

  terminate = (): void => {
    if (!this._timeout) return
    clearTimeout(this._timeout)
  }
}

class Runner<N extends WorkerMethodName> {
  constructor(methodName: N, onChange: ChangeCallback<N>) {
    this.methodName = methodName
    this.onChange = onChange
  }

  private methodName: N
  private worker: Worker | StubWorker<N> | null = null
  private next: NextRun<N> | null = null
  private onChange: ChangeCallback<N>
  private isBusy = false
  private loading = true

  private _initWorker = (): Worker | StubWorker<N> => {
    const w = window.Worker
      ? new Worker(`${pathPrefix}/static/worker.js`)
      : new StubWorker(this.methodName)

    w.onmessage = (e: RecvMessage<N>) => {
      this.isBusy = false
      this.onChange(e.data)
      this._send()
    }

    return w
  }

  private _send = (): void => {
    if (!this.next) {
      this.loading = false
      return
    }
    if (!this.worker) {
      this.worker = this._initWorker()
    }
    if (this.isBusy) return
    this.worker.postMessage({ method: this.methodName, arg: this.next.arg })
    this.isBusy = true
    this.next = null
  }

  run = (arg: WorkerParam<N>): void => {
    this.next = { arg }
    this.loading = true
    this._send()
  }

  shutdown = (): void => {
    if (!this.worker) return
    this.worker.terminate()
    this.worker = null
  }

  isLoading = (): boolean => {
    return this.loading
  }
}

// useWorker runs a method in a separate worker context
// when debugging, be sure to switch to the worker.js context, or
// disable "Selected context only" under the devtools Console settings
export function useWorker<N extends WorkerMethodName>(
  methodName: N,
  methodOpts: WorkerParam<N>,
  defaultValue: WorkerResult<N>,
): WorkerReturnType<N> {
  if (!(methodName in methods)) {
    throw new Error(`method must be a valid method from app/worker/methods.ts`)
  }

  const [result, setResult] = useState(defaultValue)
  const [worker, setWorker] = useState<Runner<N> | null>(null)

  useEffect(() => {
    const w = new Runner<N>(methodName, setResult)
    setWorker(w)
    return w.shutdown
  }, [])

  useMemo(() => {
    if (!worker) return
    worker.run(methodOpts)
  }, [worker, methodOpts])

  const loadingStatus = worker?.isLoading() || false

  return [result, { loading: loadingStatus }]
}
