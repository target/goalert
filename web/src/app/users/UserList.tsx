import React, { useState } from 'react'
import { gql } from '@apollo/client'

import { UserAvatar } from '../util/avatars'
import QueryList from '../lists/QueryList'
import UserPhoneNumberFilterContainer from './UserPhoneNumberFilterContainer'
import CreateFAB from '../lists/CreateFAB'
import UserCreateDialog from './UserCreateDialog'
import { useConfigValue, useSessionInfo } from '../util/RequireConfig'

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
  const [authDisableBasic] = useConfigValue('Auth.DisableBasic')

  return (
    <React.Fragment>
      <QueryList
        query={query}
        mapDataNode={(n) => ({
          title: n.name,
          subText: n.email,
          url: n.id,
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
      />
      {ready && isAdmin && !authDisableBasic && (
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
