import React, { useState } from 'react'
import { useQuery } from 'urql'
import Button from '@mui/material/Button'
import Tooltip from '@mui/material/Tooltip'

import CalendarSubscribeCreateDialog from './CalendarSubscribeCreateDialog'
import { calendarSubscriptionsQuery } from '../../users/UserCalendarSubscriptionList'
import { useConfigValue, useSessionInfo } from '../../util/RequireConfig'
import _ from 'lodash'
import { UserCalendarSubscription } from '../../../schema'

interface CalendarSubscribeButtonProps {
  scheduleID: string
}

export default function CalendarSubscribeButton({
  scheduleID,
}: CalendarSubscribeButtonProps): JSX.Element {
  const [creationDisabled] = useConfigValue(
    'General.DisableCalendarSubscriptions',
  )

  const [showDialog, setShowDialog] = useState(false)
  const { userID } = useSessionInfo()

  const [{ data, error }] = useQuery({
    query: calendarSubscriptionsQuery,
    variables: {
      id: userID,
    },
    pause: !userID,
  })

  const numSubs = _.get(data, 'user.calendarSubscriptions', []).filter(
    (cs: UserCalendarSubscription) =>
      cs.scheduleID === scheduleID && !cs.disabled,
  ).length

  let context =
    'Subscribe to your personal shifts from your preferred calendar app'
  if (!error && numSubs > 0) {
    context = `You have ${numSubs} active subscription${
      numSubs > 1 ? 's' : ''
    } for this schedule`
  } else if (creationDisabled) {
    context =
      'Creating subscriptions is currently disabled by your administrator'
  }

  return (
    <React.Fragment>
      <Tooltip
        title={context}
        placement='top-start'
        PopperProps={{
          // @ts-expect-error data-cy is not a valid prop
          'data-cy': 'subscribe-btn-context',
        }}
      >
        <Button
          data-cy='subscribe-btn'
          aria-label='Subscribe to this schedule'
          disabled={Boolean(creationDisabled)}
          onClick={() => setShowDialog(true)}
          variant='contained'
        >
          Subscribe
        </Button>
      </Tooltip>

      {showDialog && (
        <CalendarSubscribeCreateDialog
          onClose={() => setShowDialog(false)}
          scheduleID={scheduleID}
        />
      )}
    </React.Fragment>
  )
}
