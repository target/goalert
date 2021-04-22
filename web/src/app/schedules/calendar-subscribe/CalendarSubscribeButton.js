import React, { useState } from 'react'
import { useQuery } from '@apollo/client'
import { PropTypes as p } from 'prop-types'
import Button from '@material-ui/core/Button'
import Tooltip from '@material-ui/core/Tooltip'
import { makeStyles } from '@material-ui/core/styles'

import CalendarSubscribeCreateDialog from './CalendarSubscribeCreateDialog'
import { calendarSubscriptionsQuery } from '../../users/UserCalendarSubscriptionList'
import { useConfigValue, useSessionInfo } from '../../util/RequireConfig'
import _ from 'lodash'

const useStyles = makeStyles((theme) => ({
  container: {
    padding: 8,
  },
  calIcon: {
    marginRight: theme.spacing(1),
  },
  captionContainer: {
    display: 'grid',
  },
}))

export default function CalendarSubscribeButton(props) {
  const [creationDisabled] = useConfigValue(
    'General.DisableCalendarSubscriptions',
  )

  const [showDialog, setShowDialog] = useState(false)
  const classes = useStyles()
  const { userID, ready } = useSessionInfo()

  const { data, error } = useQuery(calendarSubscriptionsQuery, {
    variables: {
      id: userID,
    },
    skip: !ready,
  })

  const numSubs = _.get(data, 'user.calendarSubscriptions', []).filter(
    (cs) => cs.scheduleID === props.scheduleID && !cs.disabled,
  ).length

  let caption =
    'Subscribe to your personal shifts from your preferred external calendar app'
  if (!error && numSubs > 0) {
    caption = `You have ${numSubs} active subscription${
      numSubs > 1 ? 's' : ''
    } for this schedule`
  } else if (creationDisabled) {
    caption =
      'Creating subscriptions is currently disabled by your administrator'
  }

  return (
    <React.Fragment>
      <div className={classes.container}>
        <Tooltip title={caption} placement='top-start' interactive>
          <Button
            data-cy='subscribe-btn'
            aria-label='Subscribe to this schedule'
            color='primary'
            disabled={creationDisabled}
            onClick={() => setShowDialog(true)}
            variant='contained'
          >
            Subscribe
          </Button>
        </Tooltip>
      </div>

      {showDialog && (
        <CalendarSubscribeCreateDialog
          onClose={() => setShowDialog(false)}
          scheduleID={props.scheduleID}
        />
      )}
    </React.Fragment>
  )
}

CalendarSubscribeButton.propTypes = {
  scheduleID: p.string,
}
