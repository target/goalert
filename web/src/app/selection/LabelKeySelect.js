import { gql } from '@apollo/client'
import { makeQuerySelect } from './QuerySelect'
import p from 'prop-types'

const query = gql`
  query($input: LabelKeySearchOptions) {
    labelKeys(input: $input) {
      nodes
    }
  }
`

export const LabelKeySelect = makeQuerySelect('LabelKeySelect', {
  query,
  mapDataNode: (key) => ({ label: key, value: key }),
})
LabelKeySelect.propTypes = {
  value: p.string,
}
