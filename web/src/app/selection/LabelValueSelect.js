import { makeQuerySelect } from './QuerySelect'
import gql from 'graphql-tag'

const query = gql`
  query($input: LabelSearchOptions) {
    labels(input: $input) {
      nodes {
        value
      }
    }
  }
`

export const LabelValueSelect = makeQuerySelect('LabelValueSelect', {
  variables: { uniqueValues: true },
  defaultQueryVariables: { uniqueValues: true },
  query,
  mapDataNode: ({ value }) => ({ label: value, value: value }),
})
