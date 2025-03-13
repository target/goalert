/* eslint-disable @typescript-eslint/no-explicit-any */
interface RawGraphQLResponse {
  data: GraphQLResponse
  errors: [any]
}

declare global {
  namespace Cypress {
    interface Chainable {
      graphql: typeof graphql
      graphqlVoid: typeof graphqlVoid
    }
  }

  interface GraphQLResponse {
    // NOTE graphql responses are arbitrary objects

    [key: string]: any
  }
}

function graphqlVoid(
  query: string,
  variables?: { [key: string]: any },
): Cypress.Chainable {
  cy.graphql(query, variables)

  return cy.then(() => {})
}

// runs a graphql query returning the data response (after asserting no errors)
function graphql(
  query: string,
  variables?: { [key: string]: any },
): Cypress.Chainable<GraphQLResponse> {
  const url = '/api/graphql'
  if (!variables) variables = {}

  return cy.request('POST', url, { query, variables }).then((res) => {
    expect(res.status, 'status code').to.eq(200)

    let data: RawGraphQLResponse
    try {
      if (typeof res.body === 'string') data = JSON.parse(res.body)
      else data = res.body
    } catch (e) {
      console.error(res.body)
      throw e
    }
    if (data.errors && data.errors[0]) {
      // causes error message to be shown
      assert.isUndefined(data.errors[0].message)
    }
    expect(data).to.not.have.property('errors')

    return data.data
  })
}

Cypress.Commands.add('graphql', graphql)
Cypress.Commands.add('graphqlVoid', graphqlVoid)

export {}
