import { useAlertCSV } from '../services/AlertMetrics/useAlertCSV'
import { useAlertMetrics } from '../services/AlertMetrics/useAlertMetrics'

const methods = {
  useAlertMetrics,
  useAlertCSV,
}
export default methods

export type WorkerMethodName = keyof typeof methods
export type WorkerMethod<N extends WorkerMethodName> = typeof methods[N]
export type WorkerResult<N extends WorkerMethodName> = ReturnType<
  WorkerMethod<N>
>
export type WorkerReturnType<N extends WorkerMethodName> = {
  result: WorkerResult<N>
  loading: boolean
}
export type WorkerParam<N extends WorkerMethodName> = Parameters<
  WorkerMethod<N>
>[0]
