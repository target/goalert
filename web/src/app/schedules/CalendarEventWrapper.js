import React, { useEffect, useState } from 'react'
import { PropTypes as p } from 'prop-types'
import Button from '@material-ui/core/Button'
import Grid from '@material-ui/core/Grid'
import Tooltip from '@material-ui/core/Tooltip'
import { makeStyles } from '@material-ui/core'
import { DateTime } from 'luxon'

const useStyles = makeStyles((theme) => ({
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
}))

export default function CalendarEventWrapper(props) {
  const classes = useStyles()
  const { children, event, onOverrideClick } = props
  const [focused, setFocused] = useState(false)

  function handleShowOverrideForm(type) {
    setFocused(false)

    onOverrideClick({
      variant: type,
      defaultValue: {
        start: event.start.toISOString(),
        end: event.end.toISOString(),
        removeUserID: event.userID,
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
  function renderInteractiveTooltip() {
    let overrideCtrls = null

    if (DateTime.fromJSDate(event.end) > DateTime.utc()) {
      overrideCtrls = (
        <React.Fragment>
          <Grid item className={classes.buttonContainer}>
            <Button
              className={classes.button}
              data-cy='replace-override'
              size='small'
              onClick={() => handleShowOverrideForm('replace')}
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
              onClick={() => handleShowOverrideForm('remove')}
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

  return (
    <Tooltip
      open={props.selected}
      classes={{
        tooltip: classes.tooltip,
        popper: classes.popper,
      }}
      interactive
      placement='bottom-start'
      PopperProps={{
        'data-cy': 'shift-tooltip',
      }}
      title={renderInteractiveTooltip()}
    >
      {React.cloneElement(children, {
        onBlur: () => setFocused(false),
        // onFocus: () => setFocused(true),
        role: 'button',
      })}
    </Tooltip>
  )
}

CalendarEventWrapper.propTypes = {
  event: p.object.isRequired,
  onOverrideClick: p.func.isRequired,
}
