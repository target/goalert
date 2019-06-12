import React from 'react'
import { PropTypes as p } from 'prop-types'
import Button from '@material-ui/core/Button'
import Grid from '@material-ui/core/Grid'
import Typography from '@material-ui/core/Typography'
import withStyles from '@material-ui/core/styles/withStyles'
import moment from 'moment'
import { connect } from 'react-redux'
import { urlParamSelector } from '../selectors'

const styles = {
  abs: {
    position: 'absolute',
  },
  container: {
    paddingBottom: '1em',
    justifyContent: 'center',
    alignItems: 'flex-end',
  },
  flexGrow: {
    flexGrow: 1,
  },
  // borderRadius: top left, top right, bottom right, bottom left
  today: {
    borderRadius: '4px 0 0 4px',
  },
  back: {
    borderRadius: 0,
    borderLeft: 0,
    borderRight: 0,
  },
  next: {
    borderRadius: '0 4px 4px 0',
  },
  month: {
    borderRadius: '4px 0 0 4px',
  },
  week: {
    borderRadius: '0 4px 4px 0',
    borderLeft: 0,
  },
}

const mapStateToProps = state => ({
  weekly: urlParamSelector(state)('weekly', false),
})

@withStyles(styles)
@connect(
  mapStateToProps,
  null,
)
export default class CalendarToolbar extends React.PureComponent {
  static propTypes = {
    date: p.instanceOf(Date).isRequired,
    label: p.string.isRequired,
    onNavigate: p.func.isRequired,
    onOverrideClick: p.func.isRequired,
    onView: p.func.isRequired,
    view: p.string.isRequired,
  }

  /*
   * Moves the calendar to the current day in
   * respect of the current view type.
   *
   * e.g. Current day: March 22, while viewing
   * April in monthly. Clicking "Today" would
   * reset the calendar back to March.
   */
  onTodayClick = e => {
    this.props.onNavigate(e, moment().toDate())
  }

  /*
   * Go backwards 1 week or 1 month, depending
   * on the current view type.
   */
  onBackClick = e => {
    const { date, weekly } = this.props
    const nextDate = weekly
      ? moment(date)
          .clone()
          .subtract(1, 'week')
      : moment(date)
          .clone()
          .subtract(1, 'month')

    this.props.onNavigate(e, nextDate.toDate())
  }

  /*
   * Advance 1 week or 1 month, depending
   * on the current view type.
   */
  onNextClick = e => {
    const { date, weekly } = this.props

    // either month or week
    let dateCopy = moment(date).clone()
    let nextDate = weekly ? dateCopy.add(1, 'week') : dateCopy.add(1, 'month')
    this.props.onNavigate(e, nextDate.toDate())
  }

  /*
   * Switches the calendar to a monthly view.
   */
  onMonthClick = () => {
    this.props.onView('month')
  }

  /*
   * Switches the calendar to a weekly view.
   */
  onWeekClick = () => {
    this.props.onView('week')
  }

  render() {
    const { classes, label, onOverrideClick, view } = this.props

    return (
      <Grid container spacing={2} className={classes.container}>
        <Grid item>
          <Button
            data-cy='show-today'
            variant='outlined'
            className={classes.today}
            onClick={this.onTodayClick}
          >
            Today
          </Button>
          <Button
            data-cy='back'
            variant='outlined'
            className={classes.back}
            onClick={this.onBackClick}
          >
            Back
          </Button>
          <Button
            data-cy='next'
            variant='outlined'
            className={classes.next}
            onClick={this.onNextClick}
          >
            Next
          </Button>
        </Grid>
        <Grid item className={classes.flexGrow} />
        <Grid item md={12} lg='auto'>
          <Button
            data-cy='show-month'
            variant='outlined'
            disabled={view === 'month'}
            className={classes.month}
            onClick={this.onMonthClick}
          >
            Month
          </Button>
          <Button
            data-cy='show-week'
            variant='outlined'
            disabled={view === 'week'}
            className={classes.week}
            onClick={this.onWeekClick}
          >
            Week
          </Button>
        </Grid>
        <Grid item md={12} lg='auto'>
          <Button
            data-cy='add-override'
            variant='outlined'
            onClick={() => onOverrideClick()}
          >
            Add Override
          </Button>
        </Grid>
        <Grid item className={classes.abs}>
          <Typography data-cy='calendar-header' variant='subtitle1'>
            {label}
          </Typography>
        </Grid>
      </Grid>
    )
  }
}
