import React, { useState } from 'react'
import { gql } from '@apollo/client'
import Grid from '@material-ui/core/Grid'

import CreateFAB from '../lists/CreateFAB'
import UserCreateDialog from './UserCreateDialog'
import { useSessionInfo } from '../util/RequireConfig'
import { usePaginatedQuery } from '../lists/usePaginatedQuery'
import { UserAvatar } from '../util/avatars'
import { User } from '../../schema'
import ControlledPaginatedList from '../lists/ControlledPaginatedList'
import UserPhoneNumberFilterContainer from './UserPhoneNumberFilterContainer'
import { useURLParam } from '../actions'

const query = gql`
  query usersQuery($input: UserSearchOptions) {
    data: users(input: $input) {
      nodes {
        id
        name
        email
      }
      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
`

function UserList(): JSX.Element {
  const [showCreateDialog, setShowCreateDialog] = useState(false)
  const { isAdmin, ready } = useSessionInfo()
  const [search] = useURLParam('search', '')
  const { q, loadMore } = usePaginatedQuery(
    query,
    search.startsWith('phone=')
      ? {
          input: {
            CMValue: search.replace(/^phone=/, ''),
            search: '',
          },
        }
      : {},
  )

  const items =
    q.data?.data?.nodes?.map((n: User) => ({
      title: n.name,
      subText: n.email,
      url: n.id,
      icon: <UserAvatar userID={n.id} />,
    })) ?? []

  return (
    <React.Fragment>
      <Grid container spacing={2}>
        <Grid item xs={12}>
          <ControlledPaginatedList
            items={items}
            loadMore={loadMore}
            isLoading={!q.data && q.loading}
            searchAdornment={<UserPhoneNumberFilterContainer />}
          />
        </Grid>
      </Grid>
      {ready && isAdmin && (
        <CreateFAB
          onClick={() => setShowCreateDialog(true)}
          title='Create User'
        />
      )}
      {showCreateDialog && (
        <UserCreateDialog onClose={() => setShowCreateDialog(false)} />
      )}
    </React.Fragment>
  )
}

export default UserList
