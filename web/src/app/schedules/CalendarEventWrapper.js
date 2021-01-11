import React, { useState } from 'react'
import { PropTypes as p } from 'prop-types'
import Button from '@material-ui/core/Button'
import Grid from '@material-ui/core/Grid'
import Popover from '@material-ui/core/Popover'
import Typography from '@material-ui/core/Typography'
import { makeStyles } from '@material-ui/core'
import { DateTime } from 'luxon'

const useStyles = makeStyles({
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
  paper: {
    padding: 8,
    maxWidth: 275,
  },
})

export default function CalendarEventWrapper(props) {
  const classes = useStyles()
  const { children, event, onOverrideClick } = props
  const [anchorEl, setAnchorEl] = useState(null)
  const open = Boolean(anchorEl)
  const id = open ? 'shift-popover' : undefined

  function handleClick(event) {
    setAnchorEl(event.currentTarget)
  }

  function handleCloseShiftInfo() {
    setAnchorEl(null)
  }

  function handleKeyDown(event) {
    const code = event.keyCode || event.which
    if (code === 13 || code === 32) {
      setAnchorEl(event.currentTarget)
    }
  }

  function handleShowOverrideForm(type) {
    handleCloseShiftInfo()

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
  function renderShiftInfo() {
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
          <Typography variant='body2'>
            {`${formatJSDate(event.start)}  –  ${formatJSDate(event.end)}`}
          </Typography>
        </Grid>
        {overrideCtrls}
      </Grid>
    )
  }

  if (!children) return null
  return (
    <React.Fragment>
      <Popover
        id={id}
        open={open}
        anchorEl={anchorEl}
        onClose={handleCloseShiftInfo}
        anchorOrigin={{
          vertical: 'bottom',
          horizontal: 'left',
        }}
        transformOrigin={{
          vertical: 'top',
          horizontal: 'left',
        }}
        classes={{
          paper: classes.paper,
        }}
      >
        {renderShiftInfo()}
      </Popover>
      {React.cloneElement(children, {
        tabIndex: 0,
        onClick: handleClick,
        onKeyDown: handleKeyDown,
        role: 'button',
        'aria-pressed': open,
        'aria-describedby': open ? id : undefined,
      })}
    </React.Fragment>
  )
}

CalendarEventWrapper.propTypes = {
  event: p.object.isRequired,
  onOverrideClick: p.func.isRequired,
}
