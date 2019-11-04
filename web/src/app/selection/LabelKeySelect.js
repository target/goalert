import { makeQuerySelect } from './QuerySelect'
import gql from 'graphql-tag'
import React from 'react'
import p from 'prop-types'

const query = gql`
  query($input: LabelKeySearchOptions) {
    labelKeys(input: $input) {
      nodes
    }
  }
`

export function LabelKeySelect(props) {
  const LabelKeySelect = makeQuerySelect('LabelKeySelect', {
    variables: { search: props.value },
    query,
    mapDataNode: key => ({ label: key, value: key }),
  })

  return <LabelKeySelect {...props} />
}

LabelKeySelect.propTypes = {
  value: p.string,
}
