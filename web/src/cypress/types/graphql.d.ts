declare namespace Cypress {
  interface Chainable {
    graphql: typeof graphql
  }
}

interface GraphQLResponse {
  // NOTE graphql responses are arbitrary objects
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  [key: string]: any
}
