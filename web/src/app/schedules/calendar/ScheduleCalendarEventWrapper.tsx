import React, { useContext, useState } from 'react'
import Button from '@mui/material/Button'
import Grid from '@mui/material/Grid'
import Popover from '@mui/material/Popover'
import Typography from '@mui/material/Typography'
import makeStyles from '@mui/styles/makeStyles'
import { DateTime } from 'luxon'
import { ScheduleCalendarContext } from '../ScheduleDetails'
import CardActions from '../../details/CardActions'
import { Edit as EditIcon, Delete as DeleteIcon } from '@mui/icons-material'
import ScheduleOverrideEditDialog from '../ScheduleOverrideEditDialog'
import ScheduleOverrideDeleteDialog from '../ScheduleOverrideDeleteDialog'
import { User } from '../../../schema'
import {
  OverrideShiftEvent,
  ScheduleCalendarEvent,
  TempSchedEvent,
  TempSchedShiftEvent,
  OnCallShiftEvent,
} from './ScheduleCalendar'

const useStyles = makeStyles({
  cardActionContainer: {
    width: '100%',
  },
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

interface ScheduleCalendarEventWrapperProps {
  event: ScheduleCalendarEvent
  children: JSX.Element
}

export default function ScheduleCalendarEventWrapper({
  children,
  event,
}: ScheduleCalendarEventWrapperProps): JSX.Element {
  const classes = useStyles()
  const [anchorEl, setAnchorEl] = useState<Element | null>(null)

  const [showEditDialog, setShowEditDialog] = useState<string | null>(null)
  const [showDeleteDialog, setShowDeleteDialog] = useState<string | null>(null)

  const { setOverrideDialog, onEditTempSched, onDeleteTempSched } = useContext(
    ScheduleCalendarContext,
  )
  const open = Boolean(anchorEl)
  const id = open ? 'shift-popover' : undefined

  function handleClick(_event: MouseEvent): void {
    setAnchorEl(_event.currentTarget as Element | null)
  }

  function handleCloseShiftInfo(): void {
    setAnchorEl(null)
  }

  function handleKeyDown(_event: KeyboardEvent): void {
    const code = _event.key
    if (code === 'Enter' || code === ' ') {
      setAnchorEl(_event.currentTarget as Element | null)
    }
  }

  function handleShowOverrideForm(_event: OnCallShiftEvent): void {
    handleCloseShiftInfo()

    setOverrideDialog({
      variantOptions: ['replace', 'remove'],
      removeUserReadOnly: true,
      defaultValue: {
        start: event.start.toISOString(),
        end: event.end.toISOString(),
        removeUserID: _event.userID,
      },
    })
  }

  function renderTempSchedButtons(
    _event: TempSchedEvent | TempSchedShiftEvent,
  ): JSX.Element {
    if (DateTime.fromJSDate(_event.end) <= DateTime.utc()) {
      // no actions on past events
      return <React.Fragment />
    }
    return (
      <React.Fragment>
        <Grid item>
          <Button
            data-cy='edit-temp-sched'
            size='small'
            onClick={() => onEditTempSched(_event.tempSched)}
            variant='contained'
            color='primary'
            title='Edit this temporary schedule'
          >
            Edit
          </Button>
        </Grid>
        <React.Fragment>
          <Grid item className={classes.flexGrow} />
          <Grid item>
            <Button
              data-cy='delete-temp-sched'
              size='small'
              onClick={() => onDeleteTempSched(_event.tempSched)}
              variant='contained'
              color='primary'
              title='Delete this temporary schedule'
            >
              Delete
            </Button>
          </Grid>
        </React.Fragment>
      </React.Fragment>
    )
  }

  function renderOverrideButtons(_event: OverrideShiftEvent): JSX.Element {
    return (
      <div className={classes.cardActionContainer}>
        <CardActions
          secondaryActions={[
            {
              icon: <EditIcon fontSize='small' />,
              label: 'Edit',
              handleOnClick: () => {
                handleCloseShiftInfo()
                setShowEditDialog(_event.override.id)
              },
            },
            {
              icon: <DeleteIcon fontSize='small' />,
              label: 'Delete',
              handleOnClick: () => {
                handleCloseShiftInfo()
                setShowDeleteDialog(_event.override.id)
              },
            },
          ]}
        />
      </div>
    )
  }

  function renderShiftButtons(_event: OnCallShiftEvent): JSX.Element {
    return (
      <React.Fragment>
        <Grid item className={classes.flexGrow} />
        <Grid item>
          <Button
            data-cy='override'
            size='small'
            onClick={() => handleShowOverrideForm(_event)}
            variant='contained'
            color='primary'
            title={`Temporarily remove ${_event.title} from this schedule`}
          >
            Override Shift
          </Button>
        </Grid>
      </React.Fragment>
    )
  }

  function renderButtons(): JSX.Element {
    if (DateTime.fromJSDate(event.end) <= DateTime.utc())
      return <React.Fragment />
    if (event.type === 'tempSched')
      return renderTempSchedButtons(event as TempSchedEvent)
    if (event.type === 'tempSchedShift')
      return renderTempSchedButtons(event as TempSchedShiftEvent)
    // if (event.fixed) return <React.Fragment /> //todo
    if (event.type === 'overrideShift')
      return renderOverrideButtons(event as OverrideShiftEvent)

    return renderShiftButtons(event as OnCallShiftEvent)
  }

  function renderOverrideDescription(_event: OverrideShiftEvent): JSX.Element {
    const getDesc = (
      addUser: User | undefined,
      removeUser: User | undefined,
    ): JSX.Element => {
      if (addUser && removeUser)
        return (
          <React.Fragment>
            <b>{addUser.name}</b> replaces <b>{removeUser.name}</b>.
          </React.Fragment>
        )
      if (addUser)
        return (
          <React.Fragment>
            Adds <b>{addUser.name}</b>.
          </React.Fragment>
        )
      if (removeUser)
        return (
          <React.Fragment>
            Removes <b>{removeUser.name}</b>.
          </React.Fragment>
        )

      return <React.Fragment />
    }

    return (
      <Grid item xs={12}>
        <Typography variant='body2'>
          {getDesc(_event.override.addUser, _event.override.removeUser)}
        </Typography>
      </Grid>
    )
  }

  /*
   * Renders an interactive tooltip when selecting
   * an event in the calendar that will show
   * the full shift start and end date times, as
   * well as the controls relevant to the event.
   */
  function renderShiftInfo(): JSX.Element {
    const fmt = (date: Date): string =>
      DateTime.fromJSDate(date).toLocaleString(DateTime.DATETIME_FULL)

    return (
      <Grid container spacing={1}>
        {event.type === 'overrideShift' &&
          renderOverrideDescription(event as OverrideShiftEvent)}
        <Grid item xs={12}>
          <Typography variant='body2'>
            {`${fmt(event.start)}  –  ${fmt(event.end)}`}
          </Typography>
        </Grid>
        {renderButtons()}
      </Grid>
    )
  }

  if (!children) return <React.Fragment />
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
          // @ts-expect-error - DOM attr for tests
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
      {showEditDialog && (
        <ScheduleOverrideEditDialog
          overrideID={showEditDialog}
          onClose={() => setShowEditDialog(null)}
        />
      )}
      {showDeleteDialog && (
        <ScheduleOverrideDeleteDialog
          overrideID={showDeleteDialog}
          onClose={() => setShowDeleteDialog(null)}
        />
      )}
    </React.Fragment>
  )
}
