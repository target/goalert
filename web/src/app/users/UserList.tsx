import React, { Suspense, useState } from 'react'
import { gql, useQuery } from 'urql'
import { UserAvatar } from '../util/avatars'
import UserPhoneNumberFilterContainer from './UserPhoneNumberFilterContainer'
import UserCreateDialog from './UserCreateDialog'
import { useSessionInfo } from '../util/RequireConfig'
import ListPageControls from '../lists/ListPageControls'
import Search from '../util/Search'
import FlatList from '../lists/FlatList'
import { UserConnection } from '../../schema'
import { useURLParam } from '../actions'
import { FavoriteIcon } from '../util/SetFavoriteButton'

const query = gql`
  query usersQuery($input: UserSearchOptions) {
    users(input: $input) {
      nodes {
        id
        name
        email
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

function UserList(): JSX.Element {
  const { isAdmin } = useSessionInfo()
  const [create, setCreate] = useState(false)
  const [search] = useURLParam<string>('search', '')
  const [pageCursors, setPageCursors] = useState(['']) // array of cursor values for previous pages
  const currentCursor = pageCursors[pageCursors.length - 1]

  const inputVars = {
    favoritesFirst: true,
    search,
    CMValue: '',
    after: currentCursor,
  }
  if (search.startsWith('phone=')) {
    inputVars.CMValue = search.replace(/^phone=/, '')
    inputVars.search = ''
  }

  const [q] = useQuery<{ users: UserConnection }>({
    query,
    variables: { input: inputVars },
    context,
  })
  const nextCursor = q.data?.users.pageInfo.hasNextPage
    ? q.data?.users.pageInfo.endCursor
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
        {create && <UserCreateDialog onClose={() => setCreate(false)} />}
      </Suspense>
      <ListPageControls
        createLabel='User'
        onBack={
          pageCursors.length > 1
            ? () => setPageCursors(pageCursors.slice(0, -1))
            : undefined
        }
        onNext={
          nextCursor
            ? () => setPageCursors([...pageCursors, nextCursor])
            : undefined
        }
        loading={q.fetching}
        onCreateClick={isAdmin ? () => setCreate(true) : undefined}
        slots={{
          search: <Search endAdornment={<UserPhoneNumberFilterContainer />} />,
          list: (
            <FlatList
              items={
                q.data?.users.nodes.map((u) => ({
                  title: u.name,
                  subText: u.email,
                  url: u.id,
                  secondaryAction: u.isFavorite ? <FavoriteIcon /> : undefined,
                  icon: <UserAvatar userID={u.id} />,
                })) || []
              }
            />
          ),
        }}
      />
    </React.Fragment>
  )
}

export default UserList
