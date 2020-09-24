import React, { Component } from 'react'
import { PropTypes as p } from 'prop-types'
import Button from '@material-ui/core/Button'
import Grid from '@material-ui/core/Grid'
import Tooltip from '@material-ui/core/Tooltip'
import withStyles from '@material-ui/core/styles/withStyles'
import { connect } from 'react-redux'
import { DateTime, Duration } from 'luxon'
import { urlParamSelector } from '../selectors'
import FixedScheduleDialog from './fixed-sched/FixedScheduleDialog'
import DeleteFixedScheduleConfirmation from './fixed-sched/DeleteFixedScheduleConfirmation'

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
    scheduleID: p.string.isRequired,
    readOnly: p.bool,
  }

  state = {
    overrideDialog: null,
    fixedSchedDialog: null,
  }

  handleShowOverrideDialog = (type) => {
    this.setState({
      overrideDialog: {
        variant: type,
        defaultValue: {
          start: this.props.event.start.toISOString(),
          end: this.props.event.end.toISOString(),
          removeUserID: this.props.event.userID,
        },
      },
    })
  }

  // handleShowFixedSchedDialog opens either an edit or a delete
  // dialog for the selected fixed shifts event
  // action: 'edit' | 'delete'
  handleShowFixedSchedDialog = (action) => {
    const { title, start, end, fixed, shifts } = this.props.event
    if (!shifts) return

    this.setState({
      fixedSchedDialog: {
        action,
        value: {
          start: DateTime.fromJSDate(start).toISO(),
          end: DateTime.fromJSDate(end).toISO(),
          shifts,
        },
      },
    })
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
    const { classes, event, readOnly } = this.props

    let actionButtons = null
    if (!readOnly && DateTime.fromJSDate(event.end) > DateTime.utc()) {
      if (event.shifts) {
        actionButtons = (
          <React.Fragment>
            <Grid item>
              <Button
                data-cy='edit-fixed-sched'
                size='small'
                onClick={() => this.handleShowFixedSchedDialog('edit')}
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
                onClick={() => this.handleShowFixedSchedDialog('delete')}
                variant='contained'
                color='primary'
                title='Delete this fixed schedule'
              >
                Delete
              </Button>
            </Grid>
          </React.Fragment>
        )
      } else if (!event.fixed) {
        actionButtons = (
          <React.Fragment>
            <Grid item>
              <Button
                data-cy='replace-override'
                size='small'
                onClick={() => this.handleShowOverrideDialog('replace')}
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
                onClick={() => this.handleShowOverrideDialog('remove')}
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
    }

    const formatJSDate = (JSDate) =>
      DateTime.fromJSDate(JSDate).toLocaleString(DateTime.DATETIME_FULL)

    return (
      <Grid container spacing={1}>
        <Grid item xs={12}>
          {`${formatJSDate(event.start)}  –  ${formatJSDate(event.end)}`}
        </Grid>
        {actionButtons}
      </Grid>
    )
  }

  render() {
    const { children, classes, readOnly, scheduleID } = this.props
    const { overrideDialog, fixedSchedDialog } = this.state

    return (
      <React.Fragment>
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
        {Boolean(overrideDialog) && !readOnly && (
          <ScheduleOverrideCreateDialog
            defaultValue={overrideDialog.defaultValue}
            variant={overrideDialog.variant}
            scheduleID={scheduleID}
            onClose={() => this.setState({ overrideDialog: null })}
            removeUserReadOnly
          />
        )}
        {fixedSchedDialog?.action === 'edit' && !readOnly && (
          <FixedScheduleDialog
            onClose={() => this.setState({ fixedSchedDialog: null })}
            scheduleID={scheduleID}
            value={fixedSchedDialog.value}
          />
        )}
        {fixedSchedDialog?.action === 'delete' && !readOnly && (
          <DeleteFixedScheduleConfirmation
            onClose={() => this.setState({ fixedSchedDialog: null })}
            scheduleID={scheduleID}
            value={fixedSchedDialog.value}
          />
        )}
      </React.Fragment>
    )
  }
}
