import { Service } from '../../../schema'

export type IntKeyMetric = {
  [type: string]: number
}
export type EPStepMetric = {
  [type: string]: number
}
export type ServiceMetrics = {
  integrationKeys: IntKeyMetric
  epSteps: EPStepMetric
  noIntKeys: Service[]
  noEPSteps: Service[]
}

export type ServiceMetricOpts = {
  services: Service[]
}

export function useServiceMetrics(opts: ServiceMetricOpts): ServiceMetrics {
  const svcsWithoutKeys: Service[] = []
  const svcsWithoutEPSteps: Service[] = []

  const svcMetrics = opts.services.reduce(
    (res, svc) => {
      if (svc.escalationPolicy?.steps.length === 0) svcsWithoutEPSteps.push(svc)
      else {
        svc.escalationPolicy?.steps.map((step) => {
          if (step.targets.length) {
            step.targets.map((tgt) => {
              res.epSteps[tgt.type] = (res.epSteps[tgt.type] || 0) + 1
            })
          }
        })
      }
      if (svc.integrationKeys.length === 0) svcsWithoutKeys.push(svc)
      else {
        svc.integrationKeys.map((key) => {
          res.integrationKeys[key.type] =
            (res.integrationKeys[key.type] || 0) + 1
        })
      }

      return res
    },
    { integrationKeys: {}, epSteps: {} } as ServiceMetrics,
  )

  return {
    ...svcMetrics,
    noIntKeys: svcsWithoutKeys,
    noEPSteps: svcsWithoutEPSteps,
  }
}
