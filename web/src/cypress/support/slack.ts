export type SlackChannel = {
  id: string
  name: string
}

function getSlackChannels(): Cypress.Chainable<SlackChannel[]> {
  return cy.graphql(`query{slackChannels{nodes{id, name}}}`).then((resp) => {
    return resp.slackChannels.nodes as SlackChannel[]
  })
}

Cypress.Commands.add('getSlackChannels', getSlackChannels)
