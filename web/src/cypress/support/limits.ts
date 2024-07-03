declare global {
  namespace Cypress {
    interface Chainable {
      getLimits: typeof getLimits
      updateLimits: typeof updateLimits
    }
  }

  interface SystemLimitInput {
    id: string
    value: number
  }
  interface SystemLimits {
    id: string
    value: number
    description: string
  }

  type Limits = Map<string, { value: number; description: string }>
}

function getLimits(): Cypress.Chainable<Limits> {
  const limits = new Map()
  const query = `query getLimits {
    systemLimits {
      id
      description
      value
    }
  }`
  return cy.graphql(query).then((res: GraphQLResponse) => {
    res.systemLimits.forEach((l: SystemLimits) => {
      limits.set(l.id, { value: l.value, description: l.description })
    })

    return limits
  })
}

function updateLimits(input: SystemLimitInput[]): Cypress.Chainable<boolean> {
  const query = `mutation updateLimits($input: [SystemLimitInput!]!){
    setSystemLimits(input: $input)
  }`

  return cy.graphql(query, { input }).then((res) => res.setSystemLimits)
}

Cypress.Commands.add('getLimits', getLimits)
Cypress.Commands.add('updateLimits', updateLimits)

export {}
