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
    }
  }

  interface Service {
    id: string
    name: string
    description: string
    is_user_favorite: boolean

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
}

function getService(svcID: string) {
  const query = `
    query GetService($id: String!) {
      service(id: $id) {
        id
        name
        description
        is_user_favorite
        epID: escalation_policy_id,
        ep: escalation_policy {
          id
          name
          description
          repeat
        }
      }
    }
  `
  return cy.graphql(query, { id: svcID }).then(res => res.service)
}

function createService(svc?: ServiceOptions): Cypress.Chainable<Service> {
  if (!svc) svc = {}
  const query = `
    mutation CreateService($input: CreateServiceInput){
      createService(input: $input) {
        id
        name
        description
        is_user_favorite
        epID: escalation_policy_id,
        ep: escalation_policy {
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
    .graphql(query, {
      input: {
        name: svc.name || 'SM Svc ' + c.word({ length: 8 }),
        description: svc.description || c.sentence(),
        escalation_policy_id: svc.epID,
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

  return cy.graphql(query, { input: { id } })
}

function createLabel(label?: LabelOptions): Cypress.Chainable<Label> {
  if (!label) label = {}
  if (!label.svcID) {
    return cy
      .createService(label.svc)
      .then(s => createLabel({ ...label, svcID: s.id }))
  }

  const query = `
    mutation SetLabel($input: SetLabelInput) {
      setLabel(input: $input)
    }
  `

  const key = label.key || `${c.word({ length: 4 })}/${c.word({ length: 3 })}`
  const value = label.value || c.word({ length: 8 })
  const svcID = label.svcID

  return cy
    .graphql(query, {
      input: {
        target_id: label.svcID,
        target_type: 'service',
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

Cypress.Commands.add('getService', getService)
Cypress.Commands.add('createService', createService)
Cypress.Commands.add('deleteService', deleteService)
Cypress.Commands.add('createLabel', createLabel)
