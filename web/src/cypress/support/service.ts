import { Chance } from 'chance'
import { IntegrationKeyType } from '../../schema'
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

      /** Creates an integration key for a given service */
      createIntKey: typeof createIntKey

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

  interface IntegrationKey {
    svcID: string
    svc: Service
    id: string
    name: string
    type: IntegrationKeyType
  }

  interface IntKeyOptions {
    svcID?: string
    svc?: ServiceOptions
    id?: string
    name?: string
    type?: IntegrationKeyType
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

function getService(svcID: string): Cypress.Chainable<Service> {
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
  return cy
    .graphql(query, { id: svcID })
    .then((res: GraphQLResponse) => res.service)
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
      .then((ep: EP) => createService({ ...svc, epID: ep.id }))
  }

  return cy
    .graphql(query, {
      input: {
        name: svc.name || 'SM Svc ' + c.word({ length: 8 }),
        description: svc.description || c.sentence(),
        escalationPolicyID: svc.epID,
        favorite: Boolean(svc.favorite),
      },
    })
    .then((res: GraphQLResponse) => res.createService)
}

function deleteService(id: string): Cypress.Chainable<void> {
  const query = `
    mutation {
      deleteService(input: $input) { id }
    }
  `
  return cy.graphqlVoid(query, { input: { id } })
}

function createLabel(label?: LabelOptions): Cypress.Chainable<Label> {
  if (!label) label = {}
  if (!label.svcID) {
    return cy
      .createService(label.svc)
      .then((s: Service) => createLabel({ ...label, svcID: s.id }))
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
    .graphql(query, {
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
    .then((svc: Service) => ({
      svcID,
      svc,
      key,
      value,
    }))
}

function createIntKey(
  intKey?: IntKeyOptions,
): Cypress.Chainable<IntegrationKey> {
  if (!intKey) intKey = {}
  if (!intKey.svcID) {
    return cy
      .createService(intKey.svc)
      .then((s: Service) => createIntKey({ svcID: s.id }))
  }

  const name = intKey.name || c.word({ length: 5 }) + ' Key'
  const svcID = intKey.svcID
  const type =
    intKey.type ||
    c.pickone([
      'email',
      'generic',
      'grafana',
      'site24x7',
      'prometheusAlertmanager',
      'universal',
    ])

  const query = `
    mutation($input: CreateIntegrationKeyInput!) {
      createIntegrationKey(input: $input) {
        id
        serviceID
        name
        type
      }
    }
  `

  return cy
    .graphql(query, {
      input: {
        serviceID: svcID,
        name,
        type,
      },
    })
    .then((res: GraphQLResponse) => {
      const key = res.createIntegrationKey
      return getService(svcID).then((svc) => {
        key.svc = svc
        return key
      })
    })
}

function createHeartbeatMonitor(
  monitor?: HeartbeatMonitorOptions,
): Cypress.Chainable<HeartbeatMonitor> {
  if (!monitor) monitor = {}
  if (!monitor.svcID) {
    return cy
      .createService(monitor.svc)
      .then((s: Service) => createHeartbeatMonitor({ svcID: s.id }))
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
    .graphql(query, {
      input: {
        serviceID: svcID,
        name,
        timeoutMinutes: timeout,
      },
    })
    .then((res: GraphQLResponse) => {
      const mon = res.createHeartbeatMonitor
      return getService(svcID).then((svc) => {
        mon.svc = svc
        return mon
      })
    })
}

Cypress.Commands.add('getService', getService)
Cypress.Commands.add('createService', createService)
Cypress.Commands.add('deleteService', deleteService)
Cypress.Commands.add('createLabel', createLabel)
Cypress.Commands.add('createIntKey', createIntKey)
Cypress.Commands.add('createHeartbeatMonitor', createHeartbeatMonitor)
