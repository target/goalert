import { gql } from '@apollo/client'
import p from 'prop-types'
import { makeQuerySelect } from './QuerySelect'

const query = gql`
  query($input: LabelValueSearchOptions) {
    labelValues(input: $input) {
      nodes
    }
  }
`

export const LabelValueSelect = makeQuerySelect('LabelValueSelect', {
  query,
  extraVariablesFunc: ({ labelKey: key, ...props }) => [props, { key }],
  mapDataNode: (value) => ({ label: value, value }),
})
LabelValueSelect.propTypes = {
  labelKey: p.string,
}
