import { Service } from '../../../schema'

export type TargetMetrics = {
  [type: string]: number
}
export type ServiceMetrics = {
  intKeyTargets: TargetMetrics
  epStepTargets: TargetMetrics
  noIntKeys: Service[]
  noEPSteps: Service[]
}

export type ServiceMetricOpts = {
  services: Service[]
}

export function useServiceMetrics(opts: ServiceMetricOpts): ServiceMetrics {
  return opts.services.reduce(
    (metrics: ServiceMetrics, svc) => {
      if (svc.escalationPolicy?.steps.length === 0) metrics.noEPSteps.push(svc)
      else {
        svc.escalationPolicy?.steps.map((step) => {
          if (step.targets.length) {
            step.targets.map((tgt) => {
              metrics.epStepTargets[tgt.type] =
                (metrics.epStepTargets[tgt.type] || 0) + 1
            })
          }
        })
      }
      if (svc.integrationKeys.length === 0) metrics.noIntKeys.push(svc)
      else {
        svc.integrationKeys.map((key) => {
          metrics.intKeyTargets[key.type] =
            (metrics.intKeyTargets[key.type] || 0) + 1
        })
      }
      return metrics
    },
    {
      intKeyTargets: {},
      epStepTargets: {},
      noIntKeys: [],
      noEPSteps: [],
    },
  )
}
