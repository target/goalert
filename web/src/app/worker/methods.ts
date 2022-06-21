import { useAlertCSV } from '../services/AlertMetrics/useAlertCSV'
import { useAlertMetrics } from '../services/AlertMetrics/useAlertMetrics'

export default {
  useAlertMetrics,
  useAlertCSV,
} as Record<string, (arg: any) => any>
