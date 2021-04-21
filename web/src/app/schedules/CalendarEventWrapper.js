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

export default function CalendarEventWrapper({
  children,
  event,
  onOverrideClick,
  onEditTempSched,
  onDeleteTempSched,
}) {
  const classes = useStyles()
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
    const code = event.key
    if (code === 'Enter' || code === ' ') {
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

  function renderTempSchedButtons() {
    return (
      <React.Fragment>
        <Grid item>
          <Button
            data-cy='edit-temp-sched'
            size='small'
            onClick={() => onEditTempSched(event.tempSched)}
            variant='contained'
            color='primary'
            title='Edit this temporary schedule'
          >
            Edit
          </Button>
        </Grid>
        {!event.isTempSchedShift && (
          <React.Fragment>
            <Grid item className={classes.flexGrow} />
            <Grid item>
              <Button
                data-cy='delete-temp-sched'
                size='small'
                onClick={() => onDeleteTempSched(event.tempSched)}
                variant='contained'
                color='primary'
                title='Delete this temporary schedule'
              >
                Delete
              </Button>
            </Grid>
          </React.Fragment>
        )}
      </React.Fragment>
    )
  }

  function renderOverrideButtons() {
    return (
      <React.Fragment>
        <Grid item>
          <Button
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
        <Grid item>
          <Button
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

  function renderButtons() {
    if (DateTime.fromJSDate(event.end) <= DateTime.utc()) return null
    if (event.tempSched) return renderTempSchedButtons()
    if (event.fixed) return null

    return renderOverrideButtons()
  }

  /*
   * Renders an interactive tooltip when selecting
   * an event in the calendar that will show
   * the full shift start and end date times, as
   * well as the controls relevant to the event.
   */
  function renderShiftInfo() {
    const fmt = (date) =>
      DateTime.fromJSDate(date).toLocaleString(DateTime.DATETIME_FULL)

    return (
      <Grid container spacing={1}>
        <Grid item xs={12}>
          <Typography variant='body2'>
            {`${fmt(event.start)}  –  ${fmt(event.end)}`}
          </Typography>
        </Grid>
        {renderButtons()}
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
        PaperProps={{
          'data-cy': 'shift-tooltip',
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
        'aria-describedby': id,
      })}
    </React.Fragment>
  )
}

CalendarEventWrapper.propTypes = {
  event: p.object.isRequired,
  onOverrideClick: p.func.isRequired,
  onEditTempSched: p.func,
  onDeleteTempSched: p.func,
}
