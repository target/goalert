import React from 'react'
import { PropTypes as p } from 'prop-types'
import { useSelector } from 'react-redux'
import {
  Button,
  ButtonGroup,
  Grid,
  makeStyles,
  Typography,
} from '@material-ui/core'
import { DateTime } from 'luxon'
import { urlParamSelector } from '../selectors'
import GroupAdd from '@material-ui/icons/GroupAdd'

const useStyles = makeStyles((theme) => ({
  tempSchedBtn: {
    marginLeft: theme.spacing(1),
  },
  container: {
    paddingBottom: '1em',
  },
  labelGridItem: {
    alignItems: 'center',
    display: 'flex',
    justifyContent: 'center',
    order: 3,
    [theme.breakpoints.up('lg')]: {
      order: 2,
    },
  },
  primaryNavBtnGroup: {
    flex: 1,
    display: 'flex',
    justifyContent: 'flex-start',
    order: 1,
  },
  secondaryBtnGroup: {
    display: 'flex',
    justifyContent: 'flex-start',
    order: 2,
    [theme.breakpoints.up('lg')]: {
      justifyContent: 'flex-end',
      order: 3,
    },
  },
}))

export default function CalendarToolbar(props) {
  const { date, onNewTempSched, view } = props
  const classes = useStyles()
  const urlParams = useSelector(urlParamSelector)
  const weekly = urlParams('weekly', false)

  const handleTodayClick = (e) => {
    props.onNavigate(e, DateTime.local().toJSDate())
  }

  const handleBackClick = (e) => {
    const timeUnit = weekly ? { weeks: 1 } : { months: 1 }
    const nextDate = DateTime.fromJSDate(date).minus(timeUnit).toJSDate()

    props.onNavigate(e, nextDate)
  }

  const handleNextClick = (e) => {
    const timeUnit = weekly ? { weeks: 1 } : { months: 1 }
    const nextDate = DateTime.fromJSDate(date).plus(timeUnit).toJSDate()

    props.onNavigate(e, nextDate)
  }

  const handleMonthViewClick = () => {
    props.onView('month')
  }

  const handleWeekViewClick = () => {
    props.onView('week')
  }

  return (
    <Grid container spacing={2} className={classes.container}>
      <Grid item xs={12} lg={4} className={classes.primaryNavBtnGroup}>
        <ButtonGroup color='primary' aria-label='Calendar Navigation'>
          <Button data-cy='show-today' onClick={handleTodayClick}>
            Today
          </Button>
          <Button data-cy='back' onClick={handleBackClick}>
            Back
          </Button>
          <Button data-cy='next' onClick={handleNextClick}>
            Next
          </Button>
        </ButtonGroup>
      </Grid>

      <Grid item xs={12} lg={4} className={classes.labelGridItem}>
        <Typography component='p' data-cy='calendar-header' variant='subtitle1'>
          {/* label is passed from react-big-calendar in <ScheduleCalendar> */}
          {props.label}
        </Typography>
      </Grid>

      <Grid item xs={12} lg={4} className={classes.secondaryBtnGroup}>
        <ButtonGroup
          color='primary'
          aria-label='Toggle between Monthly and Weekly views'
        >
          <Button
            data-cy='show-month'
            disabled={view === 'month'}
            onClick={handleMonthViewClick}
          >
            Month
          </Button>
          <Button
            data-cy='show-week'
            disabled={view === 'week'}
            onClick={handleWeekViewClick}
          >
            Week
          </Button>
        </ButtonGroup>
        {onNewTempSched && (
          <Button
            data-cy='new-temp-sched'
            variant='contained'
            size='small'
            color='primary'
            className={classes.tempSchedBtn}
            onClick={() => onNewTempSched()}
            startIcon={<GroupAdd />}
            title='Make temporary change to this schedule'
          >
            Temp Sched
          </Button>
        )}
      </Grid>
    </Grid>
  )
}

CalendarToolbar.propTypes = {
  date: p.instanceOf(Date).isRequired,
  label: p.string.isRequired,
  onNavigate: p.func.isRequired,
  onView: p.func.isRequired,
  view: p.string.isRequired,
  onNewTempSched: p.func,
}
