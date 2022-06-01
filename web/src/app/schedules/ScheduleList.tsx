import React from 'react'
import { gql } from 'urql'
import SimpleListPage from '../lists/SimpleListPage'
import ScheduleCreateDialog from './ScheduleCreateDialog'

const query = gql`
  query schedulesQuery($input: ScheduleSearchOptions) {
    data: schedules(input: $input) {
      nodes {
        id
        name
        description
        isFavorite
      }
      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
`

export default function ScheduleList(): JSX.Element {
  return (
    <SimpleListPage
      query={query}
      variables={{ input: { favoritesFirst: true } }}
      mapDataNode={(n) => ({
        title: n.name,
        subText: n.description,
        url: n.id,
        isFavorite: n.isFavorite,
      })}
      createForm={<ScheduleCreateDialog />}
      createLabel='Schedule'
    />
  )
}
