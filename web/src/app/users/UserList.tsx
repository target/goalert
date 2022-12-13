import React, { useState } from 'react'
import { gql } from '@apollo/client'
import { Button } from '@mui/material'
import { PersonAdd } from '@mui/icons-material'

import { UserAvatar } from '../util/avatars'
import QueryList from '../lists/QueryList'
import UserPhoneNumberFilterContainer from './UserPhoneNumberFilterContainer'
import CreateFAB from '../lists/CreateFAB'
import UserCreateDialog from './UserCreateDialog'
import { useSessionInfo } from '../util/RequireConfig'
import { useIsWidthDown } from '../util/useWidth'

const query = gql`
  query usersQuery($input: UserSearchOptions) {
    data: users(input: $input) {
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

function UserList(): JSX.Element {
  const { isAdmin, ready } = useSessionInfo()
  const isMobile = useIsWidthDown('md')
  const [showCreateDialog, setShowCreateDialog] = useState(false)

  return (
    <React.Fragment>
      <QueryList
        query={query}
        variables={{ input: { favoritesFirst: true } }}
        mapDataNode={(n) => ({
          title: n.name,
          subText: n.email,
          url: n.id,
          isFavorite: n.isFavorite,
          icon: <UserAvatar userID={n.id} />,
        })}
        mapVariables={(vars) => {
          if (vars?.input.search.startsWith('phone=')) {
            vars.input.CMValue = vars.input.search.replace(/^phone=/, '')
            vars.input.search = ''
          }
          return vars
        }}
        searchAdornment={<UserPhoneNumberFilterContainer />}
        CreateDialog={isAdmin ? UserCreateDialog : undefined}
        createLabel='User'
      />

      {/* rendering here instead of in QueryList since we are also checking if admin */}
      {ready && isAdmin && isMobile && (
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
