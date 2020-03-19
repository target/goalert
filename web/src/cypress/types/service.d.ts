// eslint-disable-next-line @typescript-eslint/no-unused-vars
namespace Cypress {
  interface Chainable {
    /** Gets a service with a specified ID */
    getService: typeof getService

    /**
     * Creates a new service, and escalation policy if epID is not specified
     */
    createService: typeof createService

    /** Delete the service with the specified ID */
    deleteService: typeof deleteService

    /** Creates a label for a given service */
    createLabel: typeof createLabel

    /** Creates a label for a given service */
    createHeartbeatMonitor: typeof createHeartbeatMonitor
  }
}

interface Service {
  id: string
  name: string
  description: string
  isFavorite: boolean

  /** The escalation policy ID for this Service. */
  epID: string

  /** Details for the escalation policy of this Service. */
  ep: EP
}

interface ServiceOptions {
  name?: string
  description?: string
  epID?: string
  ep?: EPOptions
  favorite?: boolean
}

interface Label {
  svcID: string
  svc: Service
  key: string
  value: string
}

interface LabelOptions {
  svcID?: string
  svc?: ServiceOptions
  key?: string
  value?: string
}

interface HeartbeatMonitor {
  svcID: string
  svc: Service
  name: string
  timeoutMinutes: number
}

interface HeartbeatMonitorOptions {
  svcID?: string
  svc?: Service
  name?: string
  timeoutMinutes?: number
}
