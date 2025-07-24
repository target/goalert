import React, { Suspense, useState } from 'react'
import { gql, useQuery } from 'urql'
import ScheduleCreateDialog from './ScheduleCreateDialog'
import { useURLParam } from '../actions'
import { ScheduleConnection } from '../../schema'
import ListPageControls from '../lists/ListPageControls'
import Search from '../util/Search'
import { FavoriteIcon } from '../util/SetFavoriteButton'
import CompList from '../lists/CompList'
import { CompListItemNav } from '../lists/CompListItems'

const query = gql`
  query schedulesQuery($input: ScheduleSearchOptions) {
    schedules(input: $input) {
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

const context = { suspense: false }
export default function ScheduleList(): JSX.Element {
  const [search] = useURLParam<string>('search', '')
  const [create, setCreate] = useState(false)
  const [cursor, setCursor] = useState('')

  const inputVars = {
    favoritesFirst: true,
    search,
    after: cursor,
  }

  const [q] = useQuery<{ schedules: ScheduleConnection }>({
    query,
    variables: { input: inputVars },
    context,
  })
  const nextCursor = q.data?.schedules.pageInfo.hasNextPage
    ? q.data?.schedules.pageInfo.endCursor
    : ''
  // cache the next page
  useQuery({
    query,
    variables: { input: { ...inputVars, after: nextCursor } },
    context,
    pause: !nextCursor,
  })

  return (
    <React.Fragment>
      <Suspense>
        {create && <ScheduleCreateDialog onClose={() => setCreate(false)} />}
      </Suspense>
      <ListPageControls
        createLabel='Schedule'
        nextCursor={nextCursor}
        onCursorChange={setCursor}
        loading={q.fetching}
        onCreateClick={() => setCreate(true)}
        slots={{
          search: <Search />,
          list: (
            <CompList emptyMessage='No results'>
              {q.data?.schedules.nodes.map((u) => (
                <CompListItemNav
                  key={u.id}
                  title={u.name}
                  subText={u.description}
                  url={u.id}
                  action={u.isFavorite ? <FavoriteIcon /> : undefined}
                />
              ))}
            </CompList>
          ),
        }}
      />
    </React.Fragment>
  )
}
