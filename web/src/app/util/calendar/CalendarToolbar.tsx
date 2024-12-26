import React from 'react'
import {
  Button,
  ButtonGroup,
  Grid,
  IconButton,
  Typography,
} from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import { Theme } from '@mui/material/styles'
import { DateTime } from 'luxon'
import { getEndOfWeek, getStartOfWeek } from '../luxon-helpers'
import { useCalendarNavigation } from './hooks'
import LeftIcon from '@mui/icons-material/ChevronLeft'
import RightIcon from '@mui/icons-material/ChevronRight'

const useStyles = makeStyles((theme: Theme) => ({
  arrowBtns: {
    marginLeft: theme.spacing(1.75),
    marginRight: theme.spacing(1.75),
  },
  container: {
    paddingBottom: theme.spacing(2),
  },
}))

type ViewType = 'month' | 'week'
interface ScheduleCalendarToolbarProps {
  filter?: React.ReactNode
  endAdornment?: React.ReactNode
}

function ScheduleCalendarToolbar(
  props: ScheduleCalendarToolbarProps,
): React.JSX.Element {
  const classes = useStyles()
  const { weekly, start, setParams: setNavParams } = useCalendarNavigation()

  const getHeader = (): string => {
    if (weekly) {
      const begin = getStartOfWeek(DateTime.fromISO(start))
      const end = getEndOfWeek(DateTime.fromISO(start))

      if (begin.month === end.month) {
        return `${end.monthLong} ${end.year}`
      }
      if (begin.year === end.year) {
        return `${begin.monthShort} — ${end.monthShort} ${end.year}`
      }

      return `${begin.monthShort} ${begin.year} — ${end.monthShort} ${end.year}`
    }

    return DateTime.fromISO(start).toFormat('LLLL yyyy')
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
    const prevStartMonth = DateTime.fromISO(start).month
    const currMonth = DateTime.now().month

    // if viewing the current month, show the current week
    if (nextView === 'week' && prevStartMonth === currMonth) {
      setNavParams({ weekly: true, start: getStartOfWeek().toISODate() })

      // if not on the current month, show the first week of the month
    } else if (nextView === 'week' && prevStartMonth !== currMonth) {
      setNavParams({
        weekly: true,
        start: DateTime.fromISO(start).startOf('month').toISODate(),
      })

      // go from week to monthly view
      // e.g. if navigating to an overlap of two months such as
      // Jan 27 - Feb 2, show the latter month (February)
    } else {
      setNavParams({
        weekly: false,
        start: getEndOfWeek(DateTime.fromISO(start))
          .startOf('month')
          .toISODate(),
      })
    }
  }

  const onNavigate = (next: DateTime): void => {
    if (weekly) {
      setNavParams({ start: getStartOfWeek(next).toISODate() })
    } else {
      setNavParams({ start: next.startOf('month').toISODate() })
    }
  }

  const handleTodayClick = (): void => {
    onNavigate(DateTime.now())
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
      justifyContent='space-between'
      alignItems='center'
    >
      <Grid item>
        <Grid container alignItems='center'>
          <Button
            data-cy='show-today'
            onClick={handleTodayClick}
            variant='outlined'
            title={DateTime.local().toFormat('cccc, LLLL d')}
          >
            Today
          </Button>

          <div className={classes.arrowBtns}>
            <IconButton
              title={`Previous ${weekly ? 'week' : 'month'}`}
              data-cy='back'
              onClick={handleBackClick}
              size='large'
            >
              <LeftIcon />
            </IconButton>
            <IconButton
              title={`Next ${weekly ? 'week' : 'month'}`}
              data-cy='next'
              onClick={handleNextClick}
              size='large'
            >
              <RightIcon />
            </IconButton>
          </div>

          <Typography component='h2' data-cy='calendar-header' variant='h5'>
            {getHeader()}
          </Typography>
        </Grid>
      </Grid>

      <Grid item>
        <Grid container alignItems='center' justifyContent='flex-end'>
          {props.filter}
          <ButtonGroup aria-label='Toggle between Monthly and Weekly views'>
            <Button
              data-cy='show-month'
              disabled={!weekly}
              onClick={() => onView('month')}
              title='Month view'
            >
              Month
            </Button>
            <Button
              data-cy='show-week'
              disabled={weekly}
              onClick={() => onView('week')}
              title='Week view'
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

export default ScheduleCalendarToolbar
