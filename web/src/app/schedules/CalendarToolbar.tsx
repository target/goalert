import React from 'react'
import {
  Button,
  ButtonGroup,
  ButtonProps,
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
  actionBtn: {
    marginLeft: theme.spacing(1),
  },
}))

type ViewType = 'month' | 'week'
interface CalendarToolbarProps {
  date: Date
  label: string
  onNavigate: (e: React.MouseEvent, date: Date) => void
  onView: (view: ViewType) => void
  view: ViewType
  actionButtonProps?: ButtonProps
}

function CalendarToolbar(props: CalendarToolbarProps): JSX.Element {
  const classes = useStyles()
  const weekly = props.view === 'week'

  return (
    <Grid container spacing={2} className={classes.container}>
      <Grid item xs={12} lg={4} className={classes.primaryNavBtnGroup}>
        <ButtonGroup color='primary' aria-label='Calendar Navigation'>
          <Button
            data-cy='show-today'
            onClick={(e) => props.onNavigate(e, DateTime.local().toJSDate())}
          >
            Today
          </Button>
          <Button
            data-cy='back'
            onClick={(e) => {
              const timeUnit = weekly ? { weeks: 1 } : { months: 1 }
              const nextDate = DateTime.fromJSDate(props.date)
                .minus(timeUnit)
                .toJSDate()

              props.onNavigate(e, nextDate)
            }}
          >
            Back
          </Button>
          <Button
            data-cy='next'
            onClick={(e) => {
              const timeUnit = weekly ? { weeks: 1 } : { months: 1 }
              const nextDate = DateTime.fromJSDate(props.date)
                .plus(timeUnit)
                .toJSDate()

              props.onNavigate(e, nextDate)
            }}
          >
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
            onClick={() => props.onView('month')}
          >
            Month
          </Button>
          <Button
            data-cy='show-week'
            disabled={props.view === 'week'}
            onClick={() => props.onView('week')}
          >
            Week
          </Button>
        </ButtonGroup>
        {props.actionButtonProps && (
          <Button
            variant='contained'
            size='small'
            color='primary'
            className={classes.actionBtn}
            {...props.actionButtonProps}
          />
        )}
      </Grid>
    </Grid>
  )
}

export default CalendarToolbar
