import React from 'react'
import { gql } from '@apollo/client'
import { Switch, Route } from 'react-router-dom'
import { UserAvatar } from '../util/avatars'
import QueryList from '../lists/QueryList'
import UserDetails from './UserDetails'
import { PageNotFound } from '../error-pages/Errors'
import { useSessionInfo } from '../util/RequireConfig'
import UserOnCallAssignmentList from './UserOnCallAssignmentList'
import Spinner from '../loading/components/Spinner'
import UserCalendarSubscriptionList from './UserCalendarSubscriptionList'
import UserSessionList from './UserSessionList'

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
      mapDataNode={(n) => ({
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

function UserCalendarSubscriptions() {
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

      <Route exact path='/profile/sessions' component={UserSessionList} />
      <Route
        exact
        path='/users/:userID/sessions'
        render={({ match }) => <UserSessionList userID={match.params.userID} />}
      />

      <Route
        exact
        path='/profile/schedule-calendar-subscriptions'
        component={UserCalendarSubscriptions}
      />
      <Route
        exact
        path='/users/:userID/schedule-calendar-subscriptions'
        render={({ match }) => (
          <UserCalendarSubscriptionList userID={match.params.userID} />
        )}
      />

      <Route component={PageNotFound} />
    </Switch>
  )
}
