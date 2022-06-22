import { useAlertCSV } from '../services/AlertMetrics/useAlertCSV'
import { useAlertMetrics } from '../services/AlertMetrics/useAlertMetrics'

const methods = {
  useAlertMetrics,
  useAlertCSV,
}
export default methods

type ValueOf<T> = T[keyof T]
export type WorkerMethod = ValueOf<typeof methods>
export type WorkerParam<M extends WorkerMethod> = Parameters<M>[0]
