import p from 'prop-types'
import { makeQuerySelect } from './QuerySelect'
import gql from 'graphql-tag'

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
