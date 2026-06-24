import { gql } from 'urql'
import { makeQuerySelect } from './QuerySelect'

const query = gql`
  query ($input: SlackChannelSearchOptions) {
    slackChannels(input: $input) {
      nodes {
        id
        name
      }
    }
  }
`

const valueQuery = gql`
  query ($id: ID!) {
    slackChannel(id: $id) {
      id
      name
    }
  }
`

export const SlackChannelSelect = makeQuerySelect('SlackChannelSelect', {
  query,
  valueQuery,
})
