import { SlackChannelConnection } from '../../schema'

export type SlackChannel = {
  id: string
  name: string
}

function getSlackChannels(): Cypress.Chainable<SlackChannel[]> {
  return cy
    .graphql(`query{slackChannels{nodes{id, name}}}`)
    .then((resp: { slackChannels: SlackChannelConnection }) => {
      return resp.slackChannels.nodes
    })
}

Cypress.Commands.add('getSlackChannels', getSlackChannels)
