import React from 'react'
import { Switch, Route } from 'react-router-dom'
import gql from 'graphql-tag'
import { UserAvatar } from '../util/avatar/types'
import QueryList from '../lists/QueryList'
import UserDetails from './UserDetails'
import { PageNotFound } from '../error-pages/Errors'
import { useSessionInfo } from '../util/RequireConfig'
import UserOnCallAssignmentList from './UserOnCallAssignmentList'
import Spinner from '../loading/components/Spinner'
import UserCalendarSubscriptionList from './UserCalendarSubscriptionList'

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

function UserList() {
  return (
    <QueryList
      query={query}
      mapDataNode={n => ({
        title: n.name,
        subText: n.email,
        url: n.id,
        icon: <UserAvatar userID={n.id} />,
      })}
    />
  )
}

function UserProfile() {
  const { userID, ready } = useSessionInfo()
  if (!ready) return <Spinner />

  return <UserDetails userID={userID} />
}

function UserOnCallAssignments() {
  const { userID, ready } = useSessionInfo()
  if (!ready) return <Spinner />
  return <UserOnCallAssignmentList userID={userID} currentUser />
}

function UserOnCallSubscriptions() {
  const { userID, ready } = useSessionInfo()
  if (!ready) return <Spinner />
  return <UserCalendarSubscriptionList userID={userID} />
}

export default function UserRouter() {
  const { userID } = useSessionInfo()

  return (
    <Switch>
      <Route exact path='/users' component={UserList} />
      <Route
        exact
        path={[`/users/${userID}`, '/profile']}
        component={UserProfile}
      />
      <Route
        exact
        path='/users/:userID'
        render={({ match }) => (
          <UserDetails userID={match.params.userID} readOnly />
        )}
      />

      <Route
        exact
        path='/profile/on-call-assignments'
        component={UserOnCallAssignments}
      />
      <Route
        exact
        path='/users/:userID/on-call-assignments'
        render={({ match }) => (
          <UserOnCallAssignmentList userID={match.params.userID} />
        )}
      />

      <Route
        exact
        path='/profile/schedule-calendar-subscriptions'
        component={UserOnCallSubscriptions}
      />
      <Route
        exact
        path='/users/:userID/schedule-calendar-subscriptions'
        render={({ match }) => (
          <UserOnCallSubscriptions userID={match.params.userID} />
        )}
      />

      <Route component={PageNotFound} />
    </Switch>
  )
}
