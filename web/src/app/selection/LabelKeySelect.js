import { makeQuerySelect } from './QuerySelect'
import gql from 'graphql-tag'

const query = gql`
  query($input: LabelKeySearchOptions) {
    labelKeys(input: $input) {
      nodes
    }
  }
`

export const LabelKeySelect = makeQuerySelect('LabelKeySelect', {
  variables: {},
  defaultQueryVariables: {},
  query,
  mapDataNode: key => ({ label: key, value: key }),
})
