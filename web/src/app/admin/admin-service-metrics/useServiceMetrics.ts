import { IntegrationKeyType, Service, TargetType } from '../../../schema'

export type TargetMetrics = {
  [type in IntegrationKeyType | TargetType]: number
}
export type ServiceMetrics = {
  keyTgtTotals: TargetMetrics
  stepTgtTotals: TargetMetrics
  noIntKeys: Service[]
  noEPSteps: Service[]
}

export type ServiceMetricOpts = {
  services: Service[]
}

export function useServiceMetrics(opts: ServiceMetricOpts): ServiceMetrics {
  return opts.services.reduce(
    (metrics: ServiceMetrics, svc) => {
      // get services without any escalation policy steps
      if (svc.escalationPolicy?.steps.length === 0) metrics.noEPSteps.push(svc)
      else {
        svc.escalationPolicy?.steps.map((step) => {
          if (step.targets.length) {
            step.targets.map((tgt) => {
              // get sum of each step target across all services
              metrics.stepTgtTotals[tgt.type] =
                (metrics.stepTgtTotals[tgt.type] || 0) + 1
            })
          }
        })
      }
      // get services without integration keys
      if (svc.integrationKeys.length === 0) metrics.noIntKeys.push(svc)
      else {
        svc.integrationKeys.map((key) => {
          // get sum of each key type across all services
          metrics.keyTgtTotals[key.type] =
            (metrics.keyTgtTotals[key.type] || 0) + 1
        })
      }
      return metrics
    },
    {
      keyTgtTotals: {},
      stepTgtTotals: {},
      noIntKeys: [],
      noEPSteps: [],
    },
  )
}
