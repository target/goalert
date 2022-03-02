declare global {
  namespace Cypress {
    interface Chainable {
      /** Gets a list of slack channels */
      getSlackChannels: () => Cypress.Chainable<SlackChannel[]>
    }
  }

  interface SlackChannel {
    id: string
    name: string
  }
}

function getSlackChannels(): Cypress.Chainable<SlackChannel[]> {
  return cy
    .graphql(`query{slackChannels{nodes{id, name}}}`)
    .then((resp: GraphQLResponse) => {
      return resp.slackChannels.nodes
    })
}

Cypress.Commands.add('getSlackChannels', getSlackChannels)

export {}
