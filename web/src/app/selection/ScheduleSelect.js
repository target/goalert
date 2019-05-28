import React from 'react'

import gql from 'graphql-tag'
import QuerySelect from './QuerySelect'

const query = gql`
  query($input: ScheduleSearchOptions) {
    schedules(input: $input) {
      nodes {
        id
        name
      }
    }
  }
`

const valueQuery = gql`
  query($id: ID!) {
    schedule(id: $id) {
      id
      name
    }
  }
`
export class ScheduleSelect extends React.PureComponent {
  render() {
    return <QuerySelect {...this.props} query={query} valueQuery={valueQuery} />
  }
}
