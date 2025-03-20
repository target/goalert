import React, { Suspense, useState } from 'react'
import { useQuery, gql } from 'urql'
import { Card, Alert } from '@mui/material'
import OtherActions from '../util/OtherActions'
import CreateFAB from '../lists/CreateFAB'
import CalendarSubscribeCreateDialog from '../schedules/calendar-subscribe/CalendarSubscribeCreateDialog'
import { Warning } from '../icons'
import CalendarSubscribeDeleteDialog from '../schedules/calendar-subscribe/CalendarSubscribeDeleteDialog'
import CalendarSubscribeEditDialog from '../schedules/calendar-subscribe/CalendarSubscribeEditDialog'
import { GenericError, ObjectNotFound } from '../error-pages'
import _ from 'lodash'
import { useConfigValue } from '../util/RequireConfig'
import { UserCalendarSubscription } from '../../schema'
import { Time } from '../util/Time'
import CompList from '../lists/CompList'
import { CompListItemNav, CompListItemText } from '../lists/CompListItems'

export const calendarSubscriptionsQuery = gql`
  query calendarSubscriptions($id: ID!) {
    user(id: $id) {
      id
      calendarSubscriptions {
        id
        name
        reminderMinutes
        scheduleID
        schedule {
          name
        }
        lastAccess
        disabled
      }
    }
  }
`

export default function UserCalendarSubscriptionList(props: {
  userID: string
}): JSX.Element {
  const userID = props.userID
  const [creationDisabled] = useConfigValue(
    'General.DisableCalendarSubscriptions',
  )
  const [showCreateDialog, setShowCreateDialog] = useState(false)
  const [showEditDialogByID, setShowEditDialogByID] = useState<string | null>(
    null,
  )
  const [showDeleteDialogByID, setShowDeleteDialogByID] = useState<
    string | null
  >(null)

  const [{ data, error }] = useQuery({
    query: calendarSubscriptionsQuery,
    variables: {
      id: userID,
    },
  })

  if (error) return <GenericError error={error.message} />
  if (!_.get(data, 'user.id')) return <ObjectNotFound />

  // sort by schedule names, then subscription names
  const subs: UserCalendarSubscription[] = data.user.calendarSubscriptions
    .slice()
    .sort((a: UserCalendarSubscription, b: UserCalendarSubscription) => {
      if ((a?.schedule?.name ?? '') < (b?.schedule?.name ?? '')) return -1
      if ((a?.schedule?.name ?? '') > (b?.schedule?.name ?? '')) return 1

      if (a.name > b.name) return 1
      if (a.name < b.name) return -1
    })

  const subheaderDict: { [key: string]: boolean } = {}
  const items: React.ReactNode[] = []

  function renderOtherActions(id: string): JSX.Element {
    return (
      <OtherActions
        actions={[
          {
            label: 'Edit',
            onClick: () => setShowEditDialogByID(id),
          },
          {
            label: 'Delete',
            onClick: () => setShowDeleteDialogByID(id),
          },
        ]}
      />
    )
  }

  // push schedule names as subheaders now that the array is sorted
  subs.forEach((sub: UserCalendarSubscription) => {
    if (!subheaderDict[sub?.schedule?.name ?? '']) {
      subheaderDict[sub?.schedule?.name ?? ''] = true
      items.push(
        <CompListItemNav
          subText={sub.schedule?.name}
          url={`/schedules/${sub.scheduleID}`}
        />,
      )
    }

    // push subscriptions under relevant schedule subheaders
    items.push(
      <CompListItemText
        title={sub.name}
        subText={
          <Time
            prefix='Last sync: '
            time={sub.lastAccess}
            format='relative'
            zero='Never'
          />
        }
        action={renderOtherActions(sub.id)}
        alwaysShowIcon
        icon={sub.disabled ? <Warning message='Disabled' /> : null}
        key={sub.id}
      />,
    )
  })

  return (
    <React.Fragment>
      {creationDisabled && (
        <Alert
          data-cy='subs-disabled-warning'
          severity='warning'
          style={{ marginBottom: '1em' }}
        >
          Calendar subscriptions are currently disabled by your administrator
        </Alert>
      )}
      <Card>
        <CompList
          data-cy='calendar-subscriptions'
          note='Showing your current on-call subscriptions for all schedules'
          emptyMessage='You are not subscribed to any schedules.'
        >
          {items}
        </CompList>
      </Card>
      {!creationDisabled && (
        <CreateFAB
          title='Create Subscription'
          onClick={() => setShowCreateDialog(true)}
        />
      )}
      {showCreateDialog && (
        <CalendarSubscribeCreateDialog
          onClose={() => setShowCreateDialog(false)}
        />
      )}
      {showEditDialogByID && (
        <Suspense>
          <CalendarSubscribeEditDialog
            calSubscriptionID={showEditDialogByID}
            onClose={() => setShowEditDialogByID(null)}
          />
        </Suspense>
      )}
      {showDeleteDialogByID && (
        <CalendarSubscribeDeleteDialog
          calSubscriptionID={showDeleteDialogByID}
          onClose={() => setShowDeleteDialogByID(null)}
        />
      )}
    </React.Fragment>
  )
}
