import { Chance } from 'chance'
import { Target } from '../../schema'
const c = new Chance()

declare global {
  namespace Cypress {
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
    isFavorite: boolean
  }

  interface EPOptions {
    name?: string
    description?: string
    repeat?: number
    stepCount?: number
    favorite?: boolean
  }

  interface EPStep {
    id: string
    stepNumber: number
    delayMinutes: number
    targets: Target[]
    escalationPolicy: EP
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
        isFavorite
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
    .graphql(policyMutation, {
      input: {
        name: ep.name || 'SM EP ' + c.word({ length: 8 }),
        description: ep.description || c.sentence(),
        repeat: ep.repeat || c.integer({ min: 1, max: 5 }),
        favorite: ep.favorite || false,
      },
    })
    .then((res: GraphQLResponse) => res.createEscalationPolicy)
    .then((pol: EP) => {
      for (let i = 0; i < stepCount; i++) {
        cy.graphql(stepMutation, {
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

  return cy.graphqlVoid(mutation, {
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
      .then((ep: EP) => createEPStep({ ...step, epID: ep.id }))
  }

  return cy
    .graphql(stepMutation, {
      input: {
        escalationPolicyID: step.epID,
        delayMinutes: step.delay || c.integer({ min: 1, max: 9000 }),
        targets: step.targets || [],
      },
    })
    .then((res: GraphQLResponse) => res.createEscalationPolicyStep)
}

Cypress.Commands.add('createEP', createEP)
Cypress.Commands.add('deleteEP', deleteEP)
Cypress.Commands.add('createEPStep', createEPStep)
