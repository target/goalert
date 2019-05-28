import React from 'react'
import QuerySelect from './QuerySelect'
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

export class LabelKeySelect extends React.PureComponent {
  render() {
    return (
      <QuerySelect
        {...this.props}
        variables={{ input: { uniqueKeys: true } }}
        defaultQueryVariables={{ input: { uniqueKeys: true } }}
        mapDataNode={node => ({
          label: node.key,
          value: node.key,
        })}
        query={query}
      />
    )
  }
}
