import React, { MouseEvent } from 'react'
import {
  Button,
  ButtonGroup,
  Grid,
  makeStyles,
  Typography,
} from '@material-ui/core'
import { DateTime } from 'luxon'

const useStyles = makeStyles((theme) => ({
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

type ViewType = 'month' | 'week'
interface CalendarToolbarProps {
  date: Date
  label: string
  onNavigate: (e: React.MouseEvent, date: Date) => void
  onView: (view: ViewType) => void
  view: ViewType
  startAdornment?: React.ReactNode
  endAdornment?: React.ReactNode
}

function CalendarToolbar(props: CalendarToolbarProps): JSX.Element {
  const classes = useStyles()
  const weekly = props.view === 'week'

  const handleTodayClick = (e: MouseEvent): void => {
    props.onNavigate(e, DateTime.local().toJSDate())
  }

  const handleNextClick = (e: MouseEvent): void => {
    const timeUnit = weekly ? { weeks: 1 } : { months: 1 }
    const nextDate = DateTime.fromJSDate(props.date).plus(timeUnit).toJSDate()
    props.onNavigate(e, nextDate)
  }

  const handleBackClick = (e: MouseEvent): void => {
    const timeUnit = weekly ? { weeks: 1 } : { months: 1 }
    const nextDate = DateTime.fromJSDate(props.date).minus(timeUnit).toJSDate()
    props.onNavigate(e, nextDate)
  }

  const handleMonthViewClick = (): void => {
    props.onView('month')
  }

  const handleWeekViewClick = (): void => {
    props.onView('week')
  }

  return (
    <Grid container spacing={2} className={classes.container}>
      <Grid item xs={12} lg={4} className={classes.primaryNavBtnGroup}>
        {props.startAdornment}
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
            disabled={props.view === 'month'}
            onClick={handleMonthViewClick}
          >
            Month
          </Button>
          <Button
            data-cy='show-week'
            disabled={props.view === 'week'}
            onClick={handleWeekViewClick}
          >
            Week
          </Button>
        </ButtonGroup>
        {props.endAdornment}
      </Grid>
    </Grid>
  )
}

export default CalendarToolbar
