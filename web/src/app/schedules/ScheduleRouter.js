import React from 'react'
import { gql } from '@apollo/client'
import { Routes, Route } from 'react-router-dom'
import ScheduleCreateDialog from './ScheduleCreateDialog'
import ScheduleDetails from './ScheduleDetails'
import ScheduleOverrideList from './ScheduleOverrideList'
import ScheduleAssignedToList from './ScheduleAssignedToList'
import ScheduleShiftList from './ScheduleShiftList'
import { PageNotFound } from '../error-pages/Errors'
import ScheduleRuleList from './ScheduleRuleList'
import SimpleListPage from '../lists/SimpleListPage'
import ScheduleOnCallNotificationsList from './on-call-notifications/ScheduleOnCallNotificationsList'

const query = gql`
  query schedulesQuery($input: ScheduleSearchOptions) {
    data: schedules(input: $input) {
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
function ScheduleList() {
  return (
    <SimpleListPage
      query={query}
      variables={{ input: { favoritesFirst: true } }}
      mapDataNode={(n) => ({
        title: n.name,
        subText: n.description,
        url: n.id,
        isFavorite: n.isFavorite,
      })}
      createForm={<ScheduleCreateDialog />}
      createLabel='Schedule'
    />
  )
}

export default function ScheduleRouter() {
  return (
    <Routes>
      <Route path='/' element={<ScheduleList />} />
      <Route path=':scheduleID' element={<ScheduleDetails />} />
      <Route path=':scheduleID/assignments' element={<ScheduleRuleList />} />
      <Route
        path=':scheduleID/on-call-notifications'
        element={<ScheduleOnCallNotificationsList />}
      />
      <Route
        path=':scheduleID/escalation-policies'
        element={<ScheduleAssignedToList />}
      />
      <Route path=':scheduleID/overrides' element={<ScheduleOverrideList />} />
      <Route path=':scheduleID/shifts' element={<ScheduleShiftList />} />
      <Route element={<PageNotFound />} />
    </Routes>
  )
}
