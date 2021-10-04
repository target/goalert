import React, { useState, useCallback } from 'react'
import p from 'prop-types'
import { gql, useQuery } from '@apollo/client'
import { Redirect } from 'react-router-dom'
import _ from 'lodash'
import { Edit, Delete } from '@material-ui/icons'

import DetailsPage from '../details/DetailsPage'
import ScheduleEditDialog from './ScheduleEditDialog'
import ScheduleDeleteDialog from './ScheduleDeleteDialog'
import ScheduleCalendarQuery from './calendar/ScheduleCalendarQuery'
import { QuerySetFavoriteButton } from '../util/QuerySetFavoriteButton'
import CalendarSubscribeButton from './calendar-subscribe/CalendarSubscribeButton'
import Spinner from '../loading/components/Spinner'
import { ObjectNotFound, GenericError } from '../error-pages'
import TempSchedDialog from './temp-sched/TempSchedDialog'
import TempSchedDeleteConfirmation from './temp-sched/TempSchedDeleteConfirmation'
import { ScheduleAvatar } from '../util/avatars'
import { useConfigValue } from '../util/RequireConfig'

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

export const ScheduleCalendarContext = React.createContext({
  onNewTempSched: () => {},
  onEditTempSched: () => {},
  onDeleteTempSched: () => {},
  // ts files infer function signature, need parameter list
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  setOverrideDialog: (overrideVal) => {},
  overrideDialog: null,
})

export default function ScheduleDetails({ scheduleID }) {
  const [showEdit, setShowEdit] = useState(false)
  const [showDelete, setShowDelete] = useState(false)
  const [configTempSchedule, setConfigTempSchedule] = useState(null)
  const [deleteTempSchedule, setDeleteTempSchedule] = useState(null)

  const [slackEnabled] = useConfigValue('Slack.Enable')

  const onNewTempSched = useCallback(() => setConfigTempSchedule(true), [])
  const onEditTempSched = useCallback(setConfigTempSchedule, [])
  const onDeleteTempSched = useCallback(setDeleteTempSchedule, [])
  const [overrideDialog, setOverrideDialog] = useState(null)

  const {
    data: _data,
    loading,
    error,
  } = useQuery(query, {
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
        avatar={<ScheduleAvatar />}
        title={data.name}
        subheader={`Time Zone: ${data.timeZone || 'Loading...'}`}
        details={data.description}
        pageContent={
          <ScheduleCalendarContext.Provider
            value={{
              onNewTempSched,
              onEditTempSched,
              onDeleteTempSched,
              setOverrideDialog,
              overrideDialog,
            }}
          >
            <ScheduleCalendarQuery scheduleID={scheduleID} />
          </ScheduleCalendarContext.Provider>
        }
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
          {
            label: 'Assignments',
            url: 'assignments',
            subText: 'Manage rules for rotations and users',
          },
          {
            label: 'Escalation Policies',
            url: 'escalation-policies',
            subText: 'Find escalation policies that link to this schedule',
          },
          {
            label: 'Overrides',
            url: 'overrides',
            subText: 'Add, remove, or replace a user temporarily',
          },
          {
            label: 'Shifts',
            url: 'shifts',
            subText: 'Review a list of past and future on-call shifts',
          },

          // only slack is supported ATM, so hide the link if disabled
          slackEnabled && {
            label: 'On-Call Notifications',
            url: 'on-call-notifications',
            subText: 'Set up notifications to know who is on-call',
          },
        ]}
      />
    </React.Fragment>
  )
}

ScheduleDetails.propTypes = {
  scheduleID: p.string.isRequired,
}
