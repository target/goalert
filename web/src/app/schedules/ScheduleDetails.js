import React, { useState, useCallback } from 'react'
import p from 'prop-types'
import { gql, useQuery } from '@apollo/client'
import { Redirect } from 'react-router-dom'
import _ from 'lodash'
import { Edit, Delete, Today as SchedulesIcon } from '@material-ui/icons'

import DetailsPage from '../details/DetailsPage'
import ScheduleEditDialog from './ScheduleEditDialog'
import ScheduleDeleteDialog from './ScheduleDeleteDialog'
import ScheduleCalendarQuery from './ScheduleCalendarQuery'
import { QuerySetFavoriteButton } from '../util/QuerySetFavoriteButton'
import CalendarSubscribeButton from './calendar-subscribe/CalendarSubscribeButton'
import Spinner from '../loading/components/Spinner'
import { ObjectNotFound, GenericError } from '../error-pages'
import TempSchedDialog from './temp-sched/TempSchedDialog'
import TempSchedDeleteConfirmation from './temp-sched/TempSchedDeleteConfirmation'

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
  const [showEdit, setShowEdit] = useState(false)
  const [showDelete, setShowDelete] = useState(false)
  const [configTempSchedule, setConfigTempSchedule] = useState(null)
  const [deleteTempSchedule, setDeleteTempSchedule] = useState(null)

  const onNewTempSched = useCallback(() => setConfigTempSchedule(true), [])
  const onEditTempSched = useCallback(setConfigTempSchedule, [])
  const onDeleteTempSched = useCallback(setDeleteTempSchedule, [])

  const { data: _data, loading, error } = useQuery(query, {
    variables: { id: scheduleID },
    returnPartialData: true,
  })

  const data = _.get(_data, 'schedule', null)

  if (loading && !data?.name) return <Spinner />
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
      {configTempSchedule && (
        <TempSchedDialog
          value={configTempSchedule === true ? null : configTempSchedule}
          onClose={() => setConfigTempSchedule(null)}
          scheduleID={scheduleID}
        />
      )}
      {deleteTempSchedule && (
        <TempSchedDeleteConfirmation
          value={deleteTempSchedule}
          onClose={() => setDeleteTempSchedule(null)}
          scheduleID={scheduleID}
        />
      )}
      <DetailsPage
        title={data.name}
        details={data.description}
        thumbnail={<SchedulesIcon color='primary' />}
        headerContent={`Time Zone: ${data.timeZone || 'Loading...'}`}
        primaryActions={[
          <CalendarSubscribeButton
            key='primary-action-subscribe'
            scheduleID={scheduleID}
          />,
        ]}
        secondaryActions={[
          {
            label: 'Edit',
            icon: <Edit />,
            handleOnClick: () => setShowEdit(true),
          },
          {
            label: 'Delete',
            icon: <Delete />,
            handleOnClick: () => setShowDelete(true),
          },
          <QuerySetFavoriteButton
            key='secondary-action-favorite'
            scheduleID={scheduleID}
          />,
        ]}
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
        primaryContent={
          <ScheduleCalendarQuery
            scheduleID={scheduleID}
            onNewTempSched={onNewTempSched}
            onEditTempSched={onEditTempSched}
            onDeleteTempSched={onDeleteTempSched}
          />
        }
      />
    </React.Fragment>
  )
}

ScheduleDetails.propTypes = {
  scheduleID: p.string.isRequired,
}
