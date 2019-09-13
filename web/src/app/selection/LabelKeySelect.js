import { makeQuerySelect } from './QuerySelect'
import gql from 'graphql-tag'

const query = gql`
  query($input: LabelSearchOptions) {
    labels(input: $input) {
      nodes {
        key
      }
    }
  }
`

export const LabelKeySelect = makeQuerySelect('LabelKeySelect', {
  variables: { uniqueKeys: true },
  defaultQueryVariables: { uniqueKeys: true },
  query,
  mapDataNode: ({ key }) => ({ label: key, value: key }),
})
