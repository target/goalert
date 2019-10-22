import React from 'react'
import p from 'prop-types'
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

export function LabelValueSelect(props) {
  const { keyValue, ...selectProps } = props
  const variables = {
    search: props.keyValue + '=*',
  }

  const LabelValueSelect = makeQuerySelect('LabelValueSelect', {
    variables,
    defaultQueryVariables: variables,
    query,
    mapDataNode: ({ value }) => ({ label: value, value: value }),
  })

  return <LabelValueSelect {...selectProps} />
}

LabelValueSelect.propTypes = {
  keyValue: p.string,
}
