import React, { useState } from 'react'
import { gql, useQuery } from '@apollo/client'
import p from 'prop-types'

import { Redirect } from 'react-router-dom'
import FormControlLabel from '@material-ui/core/FormControlLabel'
import Grid from '@material-ui/core/Grid'
import Switch from '@material-ui/core/Switch'
import _ from 'lodash'

import DetailsPage from '../details/DetailsPage'
import { UserSelect } from '../selection'
import FilterContainer from '../util/FilterContainer'
import PageActions from '../util/PageActions'
import OtherActions from '../util/OtherActions'
import ScheduleEditDialog from './ScheduleEditDialog'
import ScheduleDeleteDialog from './ScheduleDeleteDialog'
import ScheduleCalendarQuery from './ScheduleCalendarQuery'
import { useURLParam, useResetURLParams } from '../actions'
import { QuerySetFavoriteButton } from '../util/QuerySetFavoriteButton'
import CalendarSubscribeButton from './calendar-subscribe/CalendarSubscribeButton'
import Spinner from '../loading/components/Spinner'
import { ObjectNotFound, GenericError } from '../error-pages'

const query = gql`
  fragment ScheduleTitleQuery on Schedule {
    id
    name
    description
  }
  query scheduleDetailsQuery($id: ID!) {
    schedule(id: $id) {
      ...ScheduleTitleQuery
      timeZone
    }
  }
`

export default function ScheduleDetails({ scheduleID }) {
  const [userFilter, setUserFilter] = useURLParam('userFilter', [])
  const [activeOnly, setActiveOnly] = useURLParam('activeOnly', false)
  const [showEdit, setShowEdit] = useState(false)
  const [showDelete, setShowDelete] = useState(false)

  const resetFilter = useResetURLParams(
    'userFilter',
    'start',
    'activeOnly',
    'tz',
    'duration',
  )

  const { data: _data, loading, error } = useQuery(query, {
    variables: { id: scheduleID },
    returnPartialData: true,
  })

  const data = _.get(_data, 'schedule', null)

  if (loading && !data) return <Spinner />
  if (error) return <GenericError error={error.message} />

  if (!data) {
    return showDelete ? <Redirect to='/schedules' push /> : <ObjectNotFound />
  }

  return (
    <React.Fragment>
      {showEdit && (
        <ScheduleEditDialog
          scheduleID={scheduleID}
          onClose={() => setShowEdit(false)}
        />
      )}
      {showDelete && (
        <ScheduleDeleteDialog
          scheduleID={scheduleID}
          onClose={() => setShowDelete(false)}
        />
      )}
      <PageActions>
        <QuerySetFavoriteButton scheduleID={scheduleID} />
        <FilterContainer onReset={resetFilter}>
          <Grid item xs={12}>
            <FormControlLabel
              control={
                <Switch
                  checked={activeOnly}
                  onChange={(e) => setActiveOnly(e.target.checked)}
                  value='activeOnly'
                />
              }
              label='Active shifts only'
            />
          </Grid>
          <Grid item xs={12}>
            <UserSelect
              label='Filter users...'
              multiple
              value={userFilter}
              onChange={setUserFilter}
            />
          </Grid>
        </FilterContainer>
        <OtherActions
          actions={[
            { label: 'Edit Schedule', onClick: () => setShowEdit(true) },
            { label: 'Delete Schedule', onClick: () => setShowDelete(true) },
          ]}
        />
      </PageActions>
      <DetailsPage
        title={data.name}
        details={data.description}
        titleFooter={
          <Grid container spacing={2}>
            <Grid item xs={12}>
              Time Zone: {data.timeZone || 'Loading...'}
            </Grid>
            <Grid item xs={12}>
              <CalendarSubscribeButton scheduleID={scheduleID} />
            </Grid>
          </Grid>
        }
        links={[
          { label: 'Assignments', url: 'assignments' },
          { label: 'Escalation Policies', url: 'escalation-policies' },
          {
            label: 'Overrides',
            url: 'overrides',
            subText: 'Temporary changes made to this schedule',
          },
          { label: 'Shifts', url: 'shifts' },
        ]}
        pageFooter={<ScheduleCalendarQuery scheduleID={scheduleID} />}
      />
    </React.Fragment>
  )
}

ScheduleDetails.propTypes = {
  scheduleID: p.string.isRequired,
}
