import { gql } from '@apollo/client'
import { makeQuerySelect } from '../selection/QuerySelect'

const query = gql`
  query notificationchannels($input: WebhookSearchOptions) {
    labelKeys(input: $input) {
      nodes
    }
  }
`

export const LabelKeySelect = makeQuerySelect('LabelKeySelect', {
  query,
  mapDataNode: (key: string) => ({ label: key, value: key }),
})
