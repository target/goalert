import { gql } from '@apollo/client'
import { makeQuerySelect } from './QuerySelect'

const query = gql`
  query ($input: ChanWebhookSearchOptions) {
    chanWebhooks(input: $input) {
      nodes {
        id
        name
      }
    }
  }
`

const valueQuery = gql`
  query ($id: String!) {
    chanWebhook(id: $id) {
      id
      name
    }
  }
`

function mapCreatedURLS(val: string): { value: string; label: string } {
  const url = new URL(val)
  console.log('in webhook select')
  return { value: val, label: url.hostname }
}

interface ChanWebhookSearchProps {
  label: string
  name: string
}

export const ChanWebhookSelect = makeQuerySelect('ChanWebhookSelect', {
  query,
  valueQuery,
  extraVariablesFunc: ({
    escalationPolicyID,
    ...props
  }: {
    escalationPolicyID: string
    props: ChanWebhookSearchProps
  }) => [props, { escalationPolicyID }],
  mapOnCreate: (val: string) => mapCreatedURLS(val),
  mapDataNode: (chanWebhook: { id: string; name: string }) => ({
    value: chanWebhook.id,
    label: chanWebhook.name,
    key: chanWebhook.id,
    subText: chanWebhook.id,
  }),
})
