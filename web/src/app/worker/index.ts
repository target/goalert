import { useEffect, useMemo, useRef, useState } from 'react'
import _ from 'lodash'
import { pathPrefix } from '../env'
import methods from './methods'

type NextRun = {
  arg: any
  method: string
}

class Runner<T, V> {
  private worker: Worker | null = null
  private next: NextRun | null = null
  private onChange: (result: V) => void = () => {}
  private isBusy: boolean = false

  private _send = () => {
    if (!this.next) return
    if (!this.worker) {
      this.worker = new Worker(`${pathPrefix}/static/worker.js`)
      this.worker.onmessage = (e) => {
        this.isBusy = false
        this.onChange(e.data)
        this._send()
      }
    }
    if (this.isBusy) return
    this.worker.postMessage(this.next)
    this.isBusy = true
    this.next = null
  }

  run = (method: string, arg: T, onChange: (result: V) => void) => {
    this.onChange = onChange
    this.next = { method, arg }
    this._send()
  }

  shutdown = () => {
    if (!this.worker) return
    this.worker.terminate()
    this.worker = null
  }
}

export function useWorker<T, V>(method: (arg: T) => V, arg: T, def: V): V {
  if (!(method.name in methods)) {
    throw new Error(`method must be a valid method from app/worker/methods.ts`)
  }

  // fallback to a simple memo if workers are unsupported
  if (!window.Worker) return useMemo(() => method(arg), [arg])

  const [result, setResult] = useState(def)
  const [worker, setWorker] = useState<Runner<T, V> | null>(null)

  useEffect(() => {
    const w = new Runner<T, V>()
    setWorker(w)
    return w.shutdown
  }, [])

  useMemo(() => {
    if (!worker) return
    worker.run(method.name, arg, setResult)
  }, [worker, arg])

  return result
}
