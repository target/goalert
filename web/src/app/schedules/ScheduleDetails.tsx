import React, { useState, useCallback } from 'react'
import { gql, useQuery } from '@apollo/client'
import { Navigate, useParams } from 'react-router-dom'
import _ from 'lodash'
import { Edit, Delete } from '@mui/icons-material'

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
import ScheduleCalendarOverrideDialog from './calendar/ScheduleCalendarOverrideDialog'
import { useIsWidthDown } from '../util/useWidth'
import { TempSchedValue } from './temp-sched/sharedUtils'

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

interface OverrideDialog {
  variantOptions: string[]
  removeUserReadOnly: boolean
  defaultValue: {
    addUserID?: string
    removeUserID?: string
    start: string
    end: string
  }
}

interface ScheduleCalendarContext {
  onNewTempSched: () => void
  onEditTempSched: (v: TempSchedValue) => void
  onDeleteTempSched: React.Dispatch<React.SetStateAction<null>>
  setOverrideDialog: React.Dispatch<React.SetStateAction<OverrideDialog | null>>
}

export const ScheduleCalendarContext =
  React.createContext<ScheduleCalendarContext>({
    onNewTempSched: () => {},
    onEditTempSched: () => {},
    onDeleteTempSched: () => {},
    setOverrideDialog: () => {},
  })

export default function ScheduleDetails(): JSX.Element {
  const { scheduleID = '' } = useParams()
  const [showEdit, setShowEdit] = useState(false)
  const [showDelete, setShowDelete] = useState(false)
  const [configTempSchedule, setConfigTempSchedule] =
    useState<Partial<TempSchedValue> | null>(null)
  const [deleteTempSchedule, setDeleteTempSchedule] = useState(null)
  const isMobile = useIsWidthDown('md')

  const [slackEnabled] = useConfigValue('Slack.Enable')

  const onNewTempSched = useCallback(() => setConfigTempSchedule({}), [])
  const onEditTempSched = useCallback(
    (v: TempSchedValue) => setConfigTempSchedule(v),
    [],
  )
  const onDeleteTempSched = useCallback(setDeleteTempSchedule, [])
  const [overrideDialog, setOverrideDialog] = useState<OverrideDialog | null>(
    null,
  )

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
    return showDelete ? <Navigate to='/schedules' /> : <ObjectNotFound />
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
          value={configTempSchedule}
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
            }}
          >
            {!isMobile && <ScheduleCalendarQuery scheduleID={scheduleID} />}
            {overrideDialog && (
              <ScheduleCalendarOverrideDialog
                defaultValue={overrideDialog.defaultValue}
                variantOptions={overrideDialog.variantOptions}
                scheduleID={scheduleID}
                onClose={() => setOverrideDialog(null)}
                removeUserReadOnly={overrideDialog.removeUserReadOnly}
              />
            )}
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
            id={scheduleID}
            type='schedule'
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
        ].concat(
          slackEnabled
            ? [
                {
                  // only slack is supported ATM, so hide the link if disabled

                  label: 'On-Call Notifications',
                  url: 'on-call-notifications',
                  subText: 'Set up notifications to know who is on-call',
                },
              ]
            : [],
        )}
      />
    </React.Fragment>
  )
}
