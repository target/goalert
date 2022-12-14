import { gql } from '@apollo/client'
import { makeQuerySelect } from './QuerySelect'

const query = gql`
  query ($input: WebhookSearchOptions) {
    webhooks(input: $input) {
      nodes {
        id
        name
      }
    }
  }
`

const valueQuery = gql`
  query ($id: ID!) {
    webhook(id: $id) {
      id
      name
    }
  }
`

export const WebhookSelect = makeQuerySelect('WebhookSelect', {
  query,
  valueQuery,
})
