import { SWOStatus } from '../../../schema'

let n = 1
let u = 1
const names: { [key: string]: string } = {}

// friendlyName will assign a persistent "friendly" name to the node.
//
// This ensures a specific ID will always refer to the same node. This
// is so that it is clear if a node disappears or a new one appears.
//
// Note: `Node 1` on one browser tab may not be the same node as `Node 1`
// on another browser tab.
export function friendlyName(id: string): string {
  if (!names[id]) {
    if (id.startsWith('unknown')) return (names[id] = 'Unknown ' + u++)
    return (names[id] = 'Node ' + n++)
  }
  return names[id]
}

export function errCheck(status: SWOStatus): string[] {
  const errs = []
  if (status.state !== 'idle')
    errs.push('Cluster is not ready, try running Reset.')

  status.nodes.forEach((node) => {
    if (node.configError)
      errs.push(`${friendlyName(node.id)} has incorrect DB URL(s).`)
    if (node.id.includes('GoAlert'))
      errs.push(
        `${friendlyName(node.id)} is a GoAlert node that is NOT in SWO mode`,
      )
  })

  return errs
}

export const toTitle = (s: string): string =>
  s.charAt(0).toUpperCase() + s.slice(1)
