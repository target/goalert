import React, { useState } from 'react'
import { PropTypes as p } from 'prop-types'
import { Button, Grid, makeStyles, Typography } from '@material-ui/core/index'
import CalendarIcon from 'mdi-material-ui/Calendar'
import CalendarSubscribeCreateDialog from './CalendarSubscribeCreateDialog'
import { Link } from 'react-router-dom'
import { useQuery } from '@apollo/react-hooks'
import { calendarSubscriptionsQuery } from '../../users/UserCalendarSubscriptionList'
import { useSessionInfo } from '../../util/RequireConfig'
import _ from 'lodash-es'

const useStyles = makeStyles(theme => ({
  calIcon: {
    marginRight: theme.spacing(1),
  },
  captionContainer: {
    display: 'grid',
  },
}))

export default function CalendarSubscribeButton(props) {
  const [showDialog, setShowDialog] = useState(false)
  const classes = useStyles()
  const { userID } = useSessionInfo()

  const { data, loading, error } = useQuery(calendarSubscriptionsQuery, {
    variables: {
      id: userID,
    },
  })

  const numSubs = _.get(data, 'user.calendarSubscriptions', []).filter(
    cs => cs.scheduleID === props.scheduleID,
  ).length

  let caption =
    'Subscribe to your shifts on this calendar from your preferred calendar app'
  if (!loading && !error && numSubs > 0) {
    if (numSubs < 99) {
      caption = `You have ${numSubs} active subscription${
        numSubs > 1 ? 's' : ''
      } for this schedule`
    } else {
      caption = 'You have 99+ active subscriptions for this schedule'
    }
  }

  return (
    <React.Fragment>
      {showDialog && (
        <CalendarSubscribeCreateDialog
          onClose={() => setShowDialog(false)}
          scheduleID={props.scheduleID}
        />
      )}
      <Grid container spacing={1}>
        <Grid item xs={12}>
          <Button
            color='primary'
            onClick={() => setShowDialog(true)}
            variant='contained'
          >
            <CalendarIcon className={classes.calIcon} />
            Subscribe
          </Button>
        </Grid>
        <Grid item xs={12} className={classes.captionContainer}>
          <Typography variant='caption' color='textSecondary'>
            {caption}
          </Typography>
          <Typography variant='caption'>
            <Link to='/profile/schedule-calendar-subscriptions'>
              Manage subscriptions
            </Link>
          </Typography>
        </Grid>
      </Grid>
    </React.Fragment>
  )
}

CalendarSubscribeButton.propTypes = {
  scheduleID: p.string,
}
