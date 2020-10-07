import React, { Component } from 'react'
import { PropTypes as p } from 'prop-types'
import Button from '@material-ui/core/Button'
import Grid from '@material-ui/core/Grid'
import Tooltip from '@material-ui/core/Tooltip'
import withStyles from '@material-ui/core/styles/withStyles'
import { connect } from 'react-redux'
import { urlParamSelector } from '../selectors'
import { DateTime, Duration } from 'luxon'

const styles = (theme) => ({
  flexGrow: {
    flexGrow: 1,
  },
  icon: {
    color: theme.palette.primary['500'],
  },
  tooltip: {
    background: theme.palette.common.white,
    color: theme.palette.text.primary,
    boxShadow: theme.shadows[1],
    fontSize: 12,
    marginTop: '0.1em',
    marginBottom: '0.1em',
  },
  popper: {
    opacity: 1,
  },
})

const mapStateToProps = (state) => {
  // false: monthly, true: weekly
  const weekly = urlParamSelector(state)('weekly', false)
  let start = urlParamSelector(state)(
    'start',
    DateTime.local().startOf('day').toISO(),
  )

  const activeOnly = urlParamSelector(state)('activeOnly', false)
  if (activeOnly) {
    start = DateTime.local().toISO()
  }

  const end = DateTime.fromISO(start)
    .plus(Duration.fromISO(weekly ? 'P7D' : 'P1M'))
    .toISO()

  return {
    start,
    end,
    userFilter: urlParamSelector(state)('userFilter', []),
    activeOnly,
  }
}

@withStyles(styles)
@connect(mapStateToProps, null)
export default class CalendarEventWrapper extends Component {
  static propTypes = {
    event: p.object.isRequired,
    onOverrideClick: p.func,
    onEditFixedSched: p.func,
    onDeleteFixedSched: p.func,
  }

  handleShowOverrideForm = (type) => {
    const { event, onOverrideClick } = this.props

    onOverrideClick({
      variant: type,
      defaultValue: {
        start: event.start.toISOString(),
        end: event.end.toISOString(),
        removeUserID: event.userID,
      },
    })
  }

  renderFixedSchedButtons() {
    const { classes, event } = this.props
    return (
      <React.Fragment>
        <Grid item>
          <Button
            data-cy='edit-fixed-sched'
            size='small'
            onClick={() => this.props.onEditFixedSched(event.fixedSched)}
            variant='contained'
            color='primary'
            title='Edit this fixed schedule'
          >
            Edit
          </Button>
        </Grid>
        <Grid item className={classes.flexGrow} />
        <Grid item>
          <Button
            data-cy='delete-fixed-sched'
            size='small'
            onClick={() => this.props.onDeleteFixedSched(event.fixedSched)}
            variant='contained'
            color='primary'
            title='Delete this fixed schedule'
          >
            Delete
          </Button>
        </Grid>
      </React.Fragment>
    )
  }

  renderOverrideButtons() {
    const { classes, event } = this.props
    return (
      <React.Fragment>
        <Grid item>
          <Button
            data-cy='replace-override'
            size='small'
            onClick={() => this.handleShowOverrideForm('replace')}
            variant='contained'
            color='primary'
            title={`Temporarily replace ${event.title} from this schedule`}
          >
            Replace
          </Button>
        </Grid>
        <Grid item className={classes.flexGrow} />
        <Grid item>
          <Button
            data-cy='remove-override'
            size='small'
            onClick={() => this.handleShowOverrideForm('remove')}
            variant='contained'
            color='primary'
            title={`Temporarily remove ${event.title} from this schedule`}
          >
            Remove
          </Button>
        </Grid>
      </React.Fragment>
    )
  }

  renderButtons() {
    const { event } = this.props
    if (DateTime.fromJSDate(event.end) <= DateTime.utc()) return null
    if (event.fixedSched) return this.renderFixedSchedButtons()
    if (event.fixed) return null

    return this.renderOverrideButtons()
  }

  /*
   * Renders an interactive tooltip when hovering
   * over an event in the calendar that will show
   * the full shift start and end date times, as
   * well as the ability to replace or remove that
   * shift as an override, if possible (not in the
   * past).
   */
  renderInteractiveTooltip = () => {
    const { event } = this.props
    const formatJSDate = (JSDate) =>
      DateTime.fromJSDate(JSDate).toLocaleString(DateTime.DATETIME_FULL)

    return (
      <Grid container spacing={1}>
        <Grid item xs={12}>
          {`${formatJSDate(event.start)}  –  ${formatJSDate(event.end)}`}
        </Grid>
        {this.renderButtons()}
      </Grid>
    )
  }

  render() {
    const { children, classes } = this.props

    return (
      <Tooltip
        classes={{
          tooltip: classes.tooltip,
          popper: classes.popper,
        }}
        interactive
        placement='bottom-start'
        PopperProps={{
          'data-cy': 'shift-tooltip',
        }}
        title={this.renderInteractiveTooltip()}
      >
        {children}
      </Tooltip>
    )
  }
}
