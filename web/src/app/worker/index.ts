import { useDeferredValue, useEffect, useMemo, useState } from 'react'
import _ from 'lodash'
import { pathPrefix } from '../env'
import methods from './methods'

interface Nameable {
  name: string
}

export function useWorker(method: string | Nameable): (arg: any) => any | null {
  const methodName = typeof method === 'string' ? method : method.name
  if (!(methodName in methods)) {
    throw new Error(`method must be a valid method from app/worker/methods.ts`)
  }
  // fallback to a simple memo if workers are unsupported
  if (!window.Worker)
    return useMemo(
      () =>
        function doWork(arg: any) {
          const _arg = useDeferredValue(arg)
          return useMemo(() => methods[methodName](_arg), [_arg])
        },
      [methodName],
    )

  const [worker, setWorker] = useState<Worker>()
  const [result, setResult] = useState(null)

  useEffect(() => {
    const w = new Worker(`${pathPrefix}/static/worker.js`)
    w.onmessage = (e) => {
      setResult(e.data)
    }
    setWorker(w)
    return () => w.terminate()
  }, [methodName])

  return useMemo(
    () =>
      function doWork(arg: any) {
        useMemo(() => {
          if (worker) worker.postMessage({ method: methodName, arg })
        }, [worker, arg])

        return result
      },
    [worker, methodName, result],
  )
}
