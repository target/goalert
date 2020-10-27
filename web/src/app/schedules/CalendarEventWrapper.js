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
  button: {
    padding: '4px',
    minHeight: 0,
    fontSize: 12,
  },
  buttonContainer: {
    display: 'flex',
    alignItems: 'center',
  },
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
    onOverrideClick: p.func.isRequired,
  }

  // Handle showing tooltips with tab focus
  // but also toggling via clicks event
  // state is tracked separately as an additional
  // click on the shift span will close the tooltip
  state = {
    eventClicked: false,
    eventHasFocus: false,
  }

  handleShowOverrideForm = (type) => {
    const { event, onOverrideClick } = this.props

    this.handleBlurTooltip()

    onOverrideClick({
      variant: type,
      defaultValue: {
        start: event.start.toISOString(),
        end: event.end.toISOString(),
        removeUserID: event.userID,
      },
    })
  }

  handleClickTooltip = () => {
    if (!this.state.eventClicked) {
      this.setState({ eventClicked: true, eventHasFocus: true })
    } else {
      this.setState({ eventClicked: false, eventHasFocus: false })
    }
  }

  handleFocusTooltip = () => {
    this.setState({ eventHasFocus: true, eventClicked: false })
  }

  handleBlurTooltip = () => {
    this.setState({ eventHasFocus: false, eventClicked: false })
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
    const { classes, event } = this.props

    let overrideCtrls = null
    if (DateTime.fromJSDate(event.end) > DateTime.utc()) {
      overrideCtrls = (
        <React.Fragment>
          <Grid item className={classes.buttonContainer}>
            <Button
              className={classes.button}
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

          <Grid item className={classes.buttonContainer}>
            <Button
              className={classes.button}
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

    const formatJSDate = (JSDate) =>
      DateTime.fromJSDate(JSDate).toLocaleString(DateTime.DATETIME_FULL)

    return (
      <Grid container spacing={1}>
        <Grid item xs={12}>
          {`${formatJSDate(event.start)}  –  ${formatJSDate(event.end)}`}
        </Grid>
        {overrideCtrls}
      </Grid>
    )
  }

  render() {
    const { children, classes } = this.props
    const { eventClicked, eventHasFocus } = this.state

    return (
      <Tooltip
        open={eventHasFocus || eventClicked}
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
        {React.cloneElement(children, {
          onClick: () => {
            children.props.onClick() // toggles selected
            this.handleClickTooltip() // toggles tooltip
          },
          onFocus: this.handleFocusTooltip,
          onBlur: this.handleBlurTooltip,
        })}
      </Tooltip>
    )
  }
}
