import React from 'react'
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

export function LabelValueSelect(props) {
  const { labelKey, ...selectProps } = props

  const LabelValueSelect = makeQuerySelect('LabelValueSelect', {
    variables: { key: labelKey },
    defaultQueryVariables: { key: labelKey },
    query,
    mapDataNode: value => ({ label: value, value }),
  })

  return <LabelValueSelect {...selectProps} />
}

LabelValueSelect.propTypes = {
  labelKey: p.string,
}
