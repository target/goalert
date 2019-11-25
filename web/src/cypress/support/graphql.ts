declare namespace Cypress {
  interface Chainable {
    graphql: typeof graphql
    graphql2: typeof graphql
  }
}

interface GraphQLResponse {
  data: any
  errors: [any]
}

function graphql2(query: string, variables?: any): Cypress.Chainable<any> {
  return graphql(query, variables, '/api/graphql')
}

// runs a graphql query returning the data response (after asserting no errors)
function graphql(
  query: string,
  variables?: any,
  url = '/v1/graphql',
): Cypress.Chainable<any> {
  if (!variables) variables = {}

  return cy.request('POST', url, { query, variables }).then(res => {
    expect(res.status, 'status code').to.eq(200)

    let data: GraphQLResponse
    try {
      if (typeof res.body === 'string') data = JSON.parse(res.body)
      else data = res.body
    } catch (e) {
      console.error(res.body)
      throw e
    }
    if (data.errors && data.errors[0]) {
      // causes error message to be shown
      expect(data.errors[0].message).to.be.empty
    }
    expect(data.errors, 'graphql errors').to.be.empty

    return data.data
  })
}

Cypress.Commands.add('graphql', graphql)
Cypress.Commands.add('graphql2', graphql2)
