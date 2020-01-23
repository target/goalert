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
import { useSelector } from 'react-redux'
import { absURLSelector } from '../../selectors'

const useStyles = makeStyles(theme => ({
  calIcon: {
    marginRight: theme.spacing(1),
  },
  captionContainer: {
    display: 'grid',
  },
}))

export default function CalendarSubscribeButton(props) {
  const absURL = useSelector(absURLSelector)
  const [showDialog, setShowDialog] = useState(false)
  const classes = useStyles()
  const { userID } = useSessionInfo()

  const { data, loading, error } = useQuery(calendarSubscriptionsQuery, {
    variables: {
      id: userID,
    },
  })

  const numSubs = _.get(data, 'user.calendarSubscriptions', []).filter(
    cs => cs.scheduleID === props.scheduleID && !cs.disabled,
  ).length

  let caption =
    'Subscribe to your shifts on this calendar from your preferred calendar app'
  if (!loading && !error && numSubs > 0) {
    caption = `You have ${numSubs} active subscription${
      numSubs > 1 ? 's' : ''
    } for this schedule`
  }

  return (
    <React.Fragment>
      <Grid container spacing={1}>
        <Grid item xs={12}>
          <Button
            data-cy='subscribe-btn'
            aria-label='Subscribe to this schedule'
            color='primary'
            onClick={() => setShowDialog(true)}
            variant='contained'
          >
            <CalendarIcon className={classes.calIcon} />
            Create Subscription
          </Button>
        </Grid>
        <Grid item xs={12} className={classes.captionContainer}>
          <Typography
            data-cy={
              loading ? 'subscribe-btn-txt-loading' : 'subscribe-btn-txt'
            }
            variant='caption'
            color='textSecondary'
          >
            {caption}
          </Typography>
          <Typography variant='caption'>
            <Link
              data-cy='manage-subscriptions-link'
              to={absURL('/profile/schedule-calendar-subscriptions')}
            >
              Manage subscriptions
            </Link>
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
