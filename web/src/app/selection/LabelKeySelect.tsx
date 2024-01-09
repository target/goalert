import { gql } from 'urql'
import { makeQuerySelect } from './QuerySelect'

const query = gql`
  query ($input: LabelKeySearchOptions) {
    labelKeys(input: $input) {
      nodes
    }
  }
`

export const LabelKeySelect = makeQuerySelect('LabelKeySelect', {
  query,
  mapDataNode: (key: string) => ({ label: key, value: key }),
})
