import React, { Suspense, useState } from 'react'
import { gql, useQuery } from 'urql'
import RotationCreateDialog from './RotationCreateDialog'
import { useURLParam } from '../actions'
import { RotationConnection } from '../../schema'
import ListPageControls from '../lists/ListPageControls'
import Search from '../util/Search'
import FlatList from '../lists/FlatList'
import { FavoriteIcon } from '../util/SetFavoriteButton'

const query = gql`
  query rotationsQuery($input: RotationSearchOptions) {
    rotations(input: $input) {
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
export default function RotationList(): JSX.Element {
  const [search] = useURLParam<string>('search', '')
  const [create, setCreate] = useState(false)
  const [cursor, setCursor] = useState('')

  const inputVars = {
    favoritesFirst: true,
    search,
    after: cursor,
  }

  const [q] = useQuery<{ rotations: RotationConnection }>({
    query,
    variables: { input: inputVars },
    context,
  })
  const nextCursor = q.data?.rotations.pageInfo.hasNextPage
    ? q.data?.rotations.pageInfo.endCursor
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
        {create && <RotationCreateDialog onClose={() => setCreate(false)} />}
      </Suspense>
      <ListPageControls
        createLabel='Rotation'
        nextCursor={nextCursor}
        onCursorChange={setCursor}
        loading={q.fetching}
        onCreateClick={() => setCreate(true)}
        slots={{
          search: <Search />,
          list: (
            <FlatList
              emptyMessage='No results'
              items={
                q.data?.rotations.nodes.map((u) => ({
                  title: u.name,
                  subText: u.description,
                  url: u.id,
                  secondaryAction: u.isFavorite ? <FavoriteIcon /> : undefined,
                })) || []
              }
            />
          ),
        }}
      />
    </React.Fragment>
  )
}
