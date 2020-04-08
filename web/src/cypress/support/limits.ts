import { SystemLimit, SystemLimitInput } from '../../schema'

export type Limits = Map<string, { value: number; description: string }>

function getLimits(): Cypress.Chainable<Limits> {
  const limits = new Map()
  const query = `query getLimits() {
    systemLimits {
      id
      description
      value
    }
  }`
  return cy.graphql2(query).then((res: GraphQLResponse) => {
    res.systemLimits.forEach((l: SystemLimit) => {
      limits.set(l.id, { value: l.value, description: l.description })
    })

    return limits
  })
}

function updateLimits(input: SystemLimitInput[]): Cypress.Chainable<boolean> {
  const query = `mutation updateLimits($input: [SystemLimitInput!]!){
    setSystemLimits(input: $input)
  }`

  return cy.graphql2(query, { input: input })
}

Cypress.Commands.add('getLimits', getLimits)
Cypress.Commands.add('updateLimits', updateLimits)
