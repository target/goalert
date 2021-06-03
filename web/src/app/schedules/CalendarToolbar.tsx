import React from 'react'
import {
  Button,
  ButtonGroup,
  Grid,
  IconButton,
  makeStyles,
  Typography,
} from '@material-ui/core'
import { DateTime } from 'luxon'
import { getEndOfWeek, getStartOfWeek } from '../util/luxon-helpers'
import { useCalendarNavigation } from './hooks'
import LeftIcon from '@material-ui/icons/ChevronLeft'
import RightIcon from '@material-ui/icons/ChevronRight'

const useStyles = makeStyles((theme) => ({
  arrowBtns: {
    marginLeft: theme.spacing(1.75),
    marginRight: theme.spacing(1.75),
  },
  container: {
    paddingBottom: theme.spacing(2),
  },
}))

type ViewType = 'month' | 'week'
interface CalendarToolbarProps {
  filter?: React.ReactNode
  endAdornment?: React.ReactNode
}

function CalendarToolbar(props: CalendarToolbarProps): JSX.Element {
  const classes = useStyles()
  const { weekly, setWeekly, start, setStart } = useCalendarNavigation()

  const getLabel = (): string => {
    if (weekly) {
      const begin = getStartOfWeek(DateTime.fromISO(start))
        .toLocal()
        .toFormat('LLLL d')
      const end = getEndOfWeek(DateTime.fromISO(start))
        .toLocal()
        .toFormat('LLLL d')
      return `${begin} — ${end}`
    }

    return DateTime.fromISO(start).toLocal().toFormat('LLLL yyyy')
  }

  /*
   * Resets the start date to the beginning of the month
   * when switching views.
   *
   * e.g. Monthly: February -> Weekly: Start at the week
   * of February 1st
   *
   * e.g. Weekly: February 17-23 -> Monthly: Start at the
   * beginning of February
   *
   * If viewing the current month however, show the current
   * week.
   */
  const onView = (nextView: ViewType): void => {
    const prevStartMonth = DateTime.fromISO(start).toLocal().month
    const currMonth = DateTime.local().month

    // if viewing the current month, show the current week
    if (nextView === 'week' && prevStartMonth === currMonth) {
      setWeekly(true)
      setStart(getStartOfWeek().toUTC().toISO())

      // if not on the current month, show the first week of the month
    } else if (nextView === 'week' && prevStartMonth !== currMonth) {
      setWeekly(true)
      setStart(
        DateTime.fromISO(start).toLocal().startOf('month').toUTC().toISO(),
      )

      // go from week to monthly view
      // e.g. if navigating to an overlap of two months such as
      // Jan 27 - Feb 2, show the latter month (February)
    } else {
      setWeekly(false)

      setStart(
        getEndOfWeek(DateTime.fromISO(start))
          .toLocal()
          .startOf('month')
          .toUTC()
          .toISO(),
      )
    }
  }

  const onNavigate = (next: DateTime): void => {
    if (weekly) {
      setStart(getStartOfWeek(next).toUTC().toISO())
    } else {
      setStart(next.toLocal().startOf('month').toUTC().toISO())
    }
  }

  const handleTodayClick = (): void => {
    onNavigate(DateTime.local())
  }

  const handleNextClick = (): void => {
    const timeUnit = weekly ? { weeks: 1 } : { months: 1 }
    const next = DateTime.fromISO(start).plus(timeUnit)
    onNavigate(next)
  }

  const handleBackClick = (): void => {
    const timeUnit = weekly ? { weeks: 1 } : { months: 1 }
    const next = DateTime.fromISO(start).minus(timeUnit)
    onNavigate(next)
  }

  return (
    <Grid
      container
      spacing={2}
      className={classes.container}
      justify='space-between'
      alignItems='center'
    >
      <Grid item>
        <Grid container alignItems='center'>
          <Button
            data-cy='show-today'
            onClick={handleTodayClick}
            variant='outlined'
            size='large'
          >
            Today
          </Button>

          <div className={classes.arrowBtns}>
            <IconButton
              title='Previous'
              data-cy='back'
              onClick={handleBackClick}
            >
              <LeftIcon />
            </IconButton>
            <IconButton title='Next' data-cy='next' onClick={handleNextClick}>
              <RightIcon />
            </IconButton>
          </div>

          <Typography component='h2' data-cy='calendar-header' variant='h5'>
            {getLabel()}
          </Typography>
        </Grid>
      </Grid>

      <Grid item>
        <Grid container alignItems='center' justify='flex-end'>
          {props.filter}
          <ButtonGroup
            color='primary'
            aria-label='Toggle between Monthly and Weekly views'
          >
            <Button
              data-cy='show-month'
              disabled={!weekly}
              onClick={() => onView('month')}
            >
              Month
            </Button>
            <Button
              data-cy='show-week'
              disabled={weekly}
              onClick={() => onView('week')}
            >
              Week
            </Button>
          </ButtonGroup>
          {props.endAdornment}
        </Grid>
      </Grid>
    </Grid>
  )
}

export default CalendarToolbar
