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
  query ($id: String!) {
    webhook(id: $id) {
      id
      name
    }
  }
`

function mapCreatedURLS(val: string): { value: string; label: string } {
  const url = new URL(val)
  return { value: val, label: url.hostname }
}

export const WebhookSelect = makeQuerySelect('WebhookSelect', {
  query,
  valueQuery,
  mapOnCreate: (val: string) => mapCreatedURLS(val),
  mapDataNode: (webhook: { id: string; name: string }) => ({
    value: webhook.id,
    label: webhook.name,
    key: webhook.id,
    subText: webhook.id,
  }),
})
