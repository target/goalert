import React from 'react'
import { gql } from '@apollo/client'
import { UserAvatar } from '../util/avatars'
import QueryList from '../lists/QueryList'
import UserPhoneNumberFilterContainer from './UserPhoneNumberFilterContainer'
import UserCreateDialog from './UserCreateDialog'
import { useSessionInfo } from '../util/RequireConfig'

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
          if (vars?.input?.search?.startsWith('phone=')) {
            vars.input.CMValue = vars.input.search.replace(/^phone=/, '')
            vars.input.search = ''
          }
          return vars
        }}
        searchAdornment={<UserPhoneNumberFilterContainer />}
        renderCreateDialog={(onClose) => {
          return <UserCreateDialog onClose={onClose} />
        }}
        createLabel='User'
        hideCreate={!ready || (ready && !isAdmin)}
      />
    </React.Fragment>
  )
}

export default UserList
