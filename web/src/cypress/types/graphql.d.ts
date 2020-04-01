// eslint-disable-next-line @typescript-eslint/no-unused-vars
namespace Cypress {
  interface Chainable {
    graphql: typeof graphql
    graphql2: typeof graphql
  }
}

interface GraphQLResponse {
  // NOTE graphql responses are arbitrary objects
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  [key: string]: any
}
