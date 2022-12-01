import { SWOStatus } from '../../../schema'

export function errCheck(status: SWOStatus): string[] {
  const errs = []
  if (status.state !== 'idle')
    errs.push('Cluster is not ready, try running Reset.')

  status.nodes.forEach((node) => {
    if (node.configError) errs.push(`Node ${node.id} has config error`)
    if (node.id.includes('GoAlert'))
      errs.push(`Node ${node.id} is a GoAlert node that is NOT in SWO mode`)
  })

  return errs
}
