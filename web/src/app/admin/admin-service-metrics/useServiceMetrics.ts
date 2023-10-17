import { IntegrationKeyType, Service, TargetType } from '../../../schema'

export type TargetMetrics = {
  [type in IntegrationKeyType | TargetType]: number
}
export type ServiceMetrics = {
  keyTgtTotals: TargetMetrics
  stepTgtTotals: TargetMetrics
  filteredServices: Service[]
}
export type ServiceMetricFilters = {
  labelKey?: string
  labelValue?: string
  epStepTgts?: string[]
  intKeyTgts?: string[]
}

export type ServiceMetricOpts = {
  services: Service[]
  filters: ServiceMetricFilters
}
export function useServiceMetrics(opts: ServiceMetricOpts): ServiceMetrics {
  const { services, filters } = opts

  const filterServices = (
    services: Service[],
    filters: ServiceMetricFilters,
  ): Service[] => {
    return services.filter((svc) => {
      if (filters.labelKey) {
        const labelMatch = svc.labels.some(
          (label) =>
            filters.labelKey === label.key &&
            (!filters.labelValue || filters.labelValue === label.value),
        )
        if (!labelMatch) return false
      }
      if (filters.epStepTgts?.length) {
        const stepTargetMatch = svc.escalationPolicy?.steps.some((step) =>
          step.targets.some((tgt) => filters.epStepTgts?.includes(tgt.type)),
        )
        if (!stepTargetMatch) return false
      }
      if (filters.intKeyTgts?.length) {
        const intKeyMatch = svc.integrationKeys.some(
          (key) => filters.intKeyTgts?.includes(key.type),
        )
        if (!intKeyMatch) return false
      }
      return true
    })
  }

  const calculateMetrics = (filteredServices: Service[]): ServiceMetrics => {
    const metrics = {
      keyTgtTotals: {},
      stepTgtTotals: {},
    } as ServiceMetrics
    filteredServices.forEach((svc) => {
      svc.escalationPolicy?.steps.forEach((step) => {
        step.targets.forEach((tgt) => {
          metrics.stepTgtTotals[tgt.type] =
            (metrics.stepTgtTotals[tgt.type] || 0) + 1
        })
      })
      svc.integrationKeys.forEach((key) => {
        metrics.keyTgtTotals[key.type] =
          (metrics.keyTgtTotals[key.type] || 0) + 1
      })
    })
    return metrics
  }

  const filteredServices = filterServices(services, filters)
  const metrics = calculateMetrics(filteredServices)
  return { ...metrics, filteredServices }
}
