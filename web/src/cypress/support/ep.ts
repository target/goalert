import { Chance } from 'chance'
const c = new Chance()

declare global {
  export namespace Cypress {
    interface Chainable {
      createEP: typeof createEP
      deleteEP: typeof deleteEP
      createEPStep: typeof createEPStep
    }
  }

  interface EP {
    id: string
    name: string
    description: string
    repeat: number
    stepCount: number
  }

  interface EPOptions {
    name?: string
    description?: string
    repeat?: number
    stepCount?: number
  }

  interface EPStep {
    delayMinutes: number
  }

  interface EPStepOptions {
    epID?: string
    ep?: EPOptions
    delay?: number
    targets?: [Target]
  }
}

const policyMutation = `
    mutation($input: CreateEscalationPolicyInput!) {
      createEscalationPolicy(input: $input) {
        id
        name
        description
        repeat
      }
    }
   `
const stepMutation = `
    mutation($input: CreateEscalationPolicyStepInput!) {
      createEscalationPolicyStep(input: $input) {
        id
        delayMinutes
        targets {
          id
          name
          type
        }
      }
    }
   `

function createEP(ep?: EPOptions): Cypress.Chainable<EP> {
  if (!ep) ep = {}
  const stepCount = ep.stepCount || 0

  return cy
    .graphql2(policyMutation, {
      input: {
        name: ep.name || 'SM EP ' + c.word({ length: 8 }),
        description: ep.description || c.sentence(),
        repeat: ep.repeat || c.integer({ min: 1, max: 5 }),
      },
    })
    .then(res => res.createEscalationPolicy)
    .then(pol => {
      for (let i = 0; i < stepCount; i++) {
        cy.graphql2(stepMutation, {
          input: {
            escalationPolicyID: pol.id,
            delayMinutes: 10,
            targets: [],
          },
        })
      }
      pol.stepCount = stepCount
      return cy.then(() => pol)
    })
}

function deleteEP(id: string): Cypress.Chainable<void> {
  const mutation = `
    mutation($input: [TargetInput!]!) {
      deleteAll(input: $input)
    }
   `

  return cy.graphql2(mutation, {
    input: [
      {
        type: 'escalationPolicy',
        id,
      },
    ],
  })
}

function createEPStep(step?: EPStepOptions): Cypress.Chainable<EPStep> {
  if (!step) step = {}

  if (!step.epID) {
    return cy
      .createEP(step.ep)
      .then(ep => createEPStep({ ...step, epID: ep.id }))
  }

  return cy
    .graphql2(stepMutation, {
      input: {
        escalationPolicyID: step.epID,
        delayMinutes: step.delay || c.integer({ min: 1, max: 9000 }),
        targets: step.targets || [],
      },
    })
    .then(res => res.createEscalationPolicyStep)
}

Cypress.Commands.add('createEP', createEP)
Cypress.Commands.add('deleteEP', deleteEP)
Cypress.Commands.add('createEPStep', createEPStep)
