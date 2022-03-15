import React from 'react'
import { Routes, Route } from 'react-router-dom'
import UserDetails from './UserDetails'
import { PageNotFound } from '../error-pages/Errors'
import { useSessionInfo } from '../util/RequireConfig'
import UserOnCallAssignmentList from './UserOnCallAssignmentList'
import Spinner from '../loading/components/Spinner'
import UserCalendarSubscriptionList from './UserCalendarSubscriptionList'
import UserSessionList from './UserSessionList'
import UserList from './UserList'

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

export function ProfileRouter() {
  return (
    <Routes>
      <Route path='/' element={<UserProfile />} />
      <Route path='/on-call-assignments' element={<UserOnCallAssignments />} />
      <Route path='/sessions' element={<UserSessionList />} />
      <Route
        path='/schedule-calendar-subscriptions'
        element={<UserCalendarSubscriptions />}
      />

      <Route element={<PageNotFound />} />
    </Routes>
  )
}

export default function UserRouter() {
  const { userID } = useSessionInfo()

  return (
    <Routes>
      <Route path='/' element={<UserList />} />
      <Route path={`/${userID}`} element={<UserProfile />} />
      <Route path=':userID' element={<UserDetails readOnly />} />
      <Route
        path=':userID/on-call-assignments'
        element={<UserOnCallAssignmentList />}
      />
      <Route path=':userID/sessions' element={<UserSessionList />} />
      <Route
        path=':userID/schedule-calendar-subscriptions'
        element={<UserCalendarSubscriptionList />}
      />

      <Route element={<PageNotFound />} />
    </Routes>
  )
}
