import React, { useState } from 'react'
import { PropTypes as p } from 'prop-types'
import { Button, Grid, makeStyles, Typography } from '@material-ui/core/index'
import CalendarIcon from 'mdi-material-ui/Calendar'
import CalendarSubscribeCreateDialog from './CalendarSubscribeCreateDialog'
import { useQuery } from '@apollo/react-hooks'
import { calendarSubscriptionsQuery } from '../../users/UserCalendarSubscriptionList'
import { useConfigValue, useSessionInfo } from '../../util/RequireConfig'
import _ from 'lodash-es'
import { AppLink } from '../../util/AppLink'

const useStyles = makeStyles((theme) => ({
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
    'Subscribe to your shifts on this calendar from your preferred calendar app'
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
      <Grid container spacing={1}>
        <Grid item xs={12}>
          <Button
            data-cy='subscribe-btn'
            aria-label='Subscribe to this schedule'
            color='primary'
            disabled={creationDisabled}
            onClick={() => setShowDialog(true)}
            variant='contained'
          >
            <CalendarIcon className={classes.calIcon} />
            Create Subscription
          </Button>
        </Grid>
        <Grid item xs={12} className={classes.captionContainer}>
          <Typography
            data-cy='subscribe-btn-txt'
            variant='caption'
            color='textSecondary'
          >
            {caption}
          </Typography>
          <Typography variant='caption'>
            <AppLink
              data-cy='manage-subscriptions-link'
              to='/profile/schedule-calendar-subscriptions'
            >
              Manage subscriptions
            </AppLink>
          </Typography>
        </Grid>
      </Grid>
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
