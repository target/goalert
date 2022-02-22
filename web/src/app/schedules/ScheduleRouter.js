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
      <Route exact path='/schedules' component={ScheduleList} />
      <Route exact path='/schedules/:scheduleID' component={ScheduleDetails} />
      <Route
        path='/schedules/:scheduleID/assignments'
        component={ScheduleRuleList}
      />
      <Route
        path='/schedules/:scheduleID/on-call-notifications'
        component={ScheduleOnCallNotificationsList}
      />
      <Route
        path='/schedules/:scheduleID/escalation-policies'
        component={ScheduleAssignedToList}
      />
      <Route
        path='/schedules/:scheduleID/overrides'
        component={ScheduleOverrideList}
      />
      <Route
        path='/schedules/:scheduleID/shifts'
        component={ScheduleShiftList}
      />
      <Route component={PageNotFound} />
    </Routes>
  )
}
