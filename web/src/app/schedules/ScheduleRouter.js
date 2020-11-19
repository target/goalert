import React from 'react'
import { gql } from '@apollo/client'
import { Switch, Route } from 'react-router-dom'
import ScheduleCreateDialog from './ScheduleCreateDialog'
import ScheduleDetails from './ScheduleDetails'
import ScheduleOverrideList from './ScheduleOverrideList'
import ScheduleAssignedToList from './ScheduleAssignedToList'
import ScheduleShiftList from './ScheduleShiftList'

import { PageNotFound } from '../error-pages/Errors'
import ScheduleRuleList from './ScheduleRuleList'
import SimpleListPage from '../lists/SimpleListPage'

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
class ScheduleList extends React.PureComponent {
  render() {
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
}

export default class ScheduleRouter extends React.PureComponent {
  render() {
    return (
      <Switch>
        <Route exact path='/schedules' component={ScheduleList} />
        <Route
          exact
          path='/schedules/:scheduleID'
          render={({ match }) => (
            <ScheduleDetails scheduleID={match.params.scheduleID} />
          )}
        />

        <Route
          path='/schedules/:scheduleID/assignments'
          render={({ match }) => (
            <ScheduleRuleList scheduleID={match.params.scheduleID} />
          )}
        />
        <Route
          path='/schedules/:scheduleID/escalation-policies'
          render={({ match }) => (
            <ScheduleAssignedToList scheduleID={match.params.scheduleID} />
          )}
        />
        <Route
          path='/schedules/:scheduleID/overrides'
          render={({ match }) => (
            <ScheduleOverrideList scheduleID={match.params.scheduleID} />
          )}
        />
        <Route
          path='/schedules/:scheduleID/shifts'
          render={({ match }) => (
            <ScheduleShiftList scheduleID={match.params.scheduleID} />
          )}
        />

        <Route component={PageNotFound} />
      </Switch>
    )
  }
}
