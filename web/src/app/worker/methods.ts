import { useAlertCSV } from '../services/AlertMetrics/useAlertCSV'
import { useAlertCountCSV } from '../admin/admin-alert-counts/useAlertCountCSV'
import { useAlertMetrics } from '../services/AlertMetrics/useAlertMetrics'
import { useAdminAlertCounts } from '../admin/admin-alert-counts/useAdminAlertCounts'

const methods = {
  useAdminAlertCounts,
  useAlertCountCSV,
  useAlertMetrics,
  useAlertCSV,
}
export default methods

export type WorkerMethodName = keyof typeof methods
export type WorkerMethod<N extends WorkerMethodName> = (typeof methods)[N]
export type WorkerResult<N extends WorkerMethodName> = ReturnType<
  WorkerMethod<N>
>
export type WorkerReturnType<N extends WorkerMethodName> = [
  result: WorkerResult<N>,
  status: { loading: boolean },
]
export type WorkerParam<N extends WorkerMethodName> = Parameters<
  WorkerMethod<N>
>[0]
