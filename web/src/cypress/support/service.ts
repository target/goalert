import { Chance } from 'chance'
const c = new Chance()

declare global {
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
}

function getService(svcID: string) {
  const query = `
    query GetService($id: ID!) {
      service(id: $id) {
        id
        name
        description
        isFavorite
        epID: escalationPolicyID,
        ep: escalationPolicy {
          id
          name
          description
          repeat
        }
      }
    }
  `
  return cy.graphql2(query, { id: svcID }).then(res => res.service)
}

function createService(svc?: ServiceOptions): Cypress.Chainable<Service> {
  if (!svc) svc = {}
  const query = `
    mutation CreateService($input: CreateServiceInput!){
      createService(input: $input) {
        id
        name
        description
        isFavorite
        epID: escalationPolicyID,
        ep: escalationPolicy {
          id
          name
          description
          repeat
        }
      }
    }
  `

  if (!svc.epID) {
    return cy
      .createEP(svc.ep)
      .then(ep => createService({ ...svc, epID: ep.id }))
  }

  return cy
    .graphql2(query, {
      input: {
        name: svc.name || 'SM Svc ' + c.word({ length: 8 }),
        description: svc.description || c.sentence(),
        escalationPolicyID: svc.epID,
        favorite: Boolean(svc.favorite),
      },
    })
    .then(res => res.createService)
}

function deleteService(id: string): Cypress.Chainable<void> {
  const query = `
    mutation {
      deleteService(input: $input) { id }
    }
  `
  return cy.graphql2(query, { input: { id } })
}

function createLabel(label?: LabelOptions): Cypress.Chainable<Label> {
  if (!label) label = {}
  if (!label.svcID) {
    return cy
      .createService(label.svc)
      .then(s => createLabel({ ...label, svcID: s.id }))
  }

  const query = `
    mutation SetLabel($input: SetLabelInput!) {
      setLabel(input: $input)
    }
  `

  const key = label.key || `${c.word({ length: 4 })}/${c.word({ length: 3 })}`
  const value = label.value || c.word({ length: 8 })
  const svcID = label.svcID

  return cy
    .graphql2(query, {
      input: {
        target: {
          type: 'service',
          id: svcID,
        },
        key,
        value,
      },
    })
    .then(() => getService(svcID))
    .then(svc => ({
      svcID,
      svc,
      key,
      value,
    }))
}

function createHeartbeatMonitor(
  monitor?: HeartbeatMonitorOptions,
): Cypress.Chainable<HeartbeatMonitor> {
  if (!monitor) monitor = {}
  if (!monitor.svcID) {
    return cy
      .createService(monitor.svc)
      .then(s => createHeartbeatMonitor({ svcID: s.id }))
  }

  const name = monitor.name || c.word({ length: 5 }) + ' Monitor'
  const timeout = monitor.timeoutMinutes || Math.trunc(Math.random() * 30) + 5
  const svcID = monitor.svcID

  const query = `
    mutation($input: CreateHeartbeatMonitorInput!) {
      createHeartbeatMonitor(input: $input) {
        id
        serviceID
        name
        timeoutMinutes
        lastState
      }
    }
  `

  return cy
    .graphql2(query, {
      input: {
        serviceID: svcID,
        name: name,
        timeoutMinutes: timeout,
      },
    })
    .then(res => {
      const mon = res.createHeartbeatMonitor
      return getService(svcID).then(svc => {
        mon.svc = svc
        return mon
      })
    })
}

Cypress.Commands.add('getService', getService)
Cypress.Commands.add('createService', createService)
Cypress.Commands.add('deleteService', deleteService)
Cypress.Commands.add('createLabel', createLabel)
Cypress.Commands.add('createHeartbeatMonitor', createHeartbeatMonitor)
