import React from 'react'

import gql from 'graphql-tag'
import QuerySelect from './QuerySelect'

const query = gql`
  query($input: TimeZoneSearchOptions) {
    timeZones(input: $input) {
      nodes {
        id
      }
    }
  }
`

export class TimeZoneSelect extends React.PureComponent {
  render() {
    return (
      <QuerySelect
        {...this.props}
        query={query}
        mapDataNode={n => ({ label: n.id, value: n.id })}
      />
    )
  }
}
