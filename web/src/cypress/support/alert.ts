import { Chance } from 'chance'
const c = new Chance()

declare global {
  namespace Cypress {
    interface Chainable {
      createAlert: typeof createAlert
      closeAlert: typeof closeAlert
    }
  }
  interface Alert {
    number: number
    id: number
    summary: string
    details: string
    serviceID: string
    service: Service
  }

  interface AlertOptions {
    summary?: string
    details?: string
    serviceID?: string

    service?: ServiceOptions
  }
}

function createAlert(a?: AlertOptions): Cypress.Chainable<Alert> {
  if (!a) a = {}
  const query = `mutation CreateAlert($input: CreateAlertInput){
      createAlert(input: $input) {
        number: _id, id, summary, details, serviceID: service_id,
        service {
          id, name, description
          epID: escalation_policy_id,
          ep: escalation_policy {
              id
              name
              description
              repeat
          }
        }
      }
    }`

  if (!a.serviceID) {
    return cy
      .createService(a.service)
      .then(svc => createAlert({ ...a, serviceID: svc.id }))
  }

  return cy
    .graphql(query, {
      input: {
        service_id: a.serviceID,
        summary: a.summary || c.sentence({ words: 3 }),
        details: a.details || c.sentence({ words: 5 }),
      },
    })
    .then(res => res.createAlert)
}

function closeAlert(id: number): Cypress.Chainable<Alert> {
  const query = `
    mutation {
      updateAlertStatus(input: $input) { id }
    }
  `

  return cy.graphql(query, { input: { id } })
}

Cypress.Commands.add('createAlert', createAlert)
Cypress.Commands.add('closeAlert', closeAlert)
