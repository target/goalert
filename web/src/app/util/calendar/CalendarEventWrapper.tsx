import React, {
  useContext,
  useState,
  MouseEvent,
  KeyboardEvent,
  Suspense,
} from 'react'
import Button from '@mui/material/Button'
import Grid from '@mui/material/Grid'
import Popover from '@mui/material/Popover'
import Typography from '@mui/material/Typography'
import makeStyles from '@mui/styles/makeStyles'
import { DateTime } from 'luxon'
import { OverrideDialogContext } from '../../schedules/ScheduleDetails'
import CardActions from '../../details/CardActions'
import { Edit as EditIcon, Delete as DeleteIcon } from '@mui/icons-material'
import ScheduleOverrideEditDialog from '../../schedules/ScheduleOverrideEditDialog'
import ScheduleOverrideDeleteDialog from '../../schedules/ScheduleOverrideDeleteDialog'
import { User } from '../../../schema'
import {
  OverrideEvent,
  ScheduleCalendarEvent,
  TempSchedEvent,
  TempSchedShiftEvent,
  OnCallShiftEvent,
} from './Calendar'
import AppLink from '../AppLink'

const useStyles = makeStyles({
  cardActionContainer: {
    width: '100%',
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
  children: React.JSX.Element
  event: ScheduleCalendarEvent
  showScheduleLink?: boolean
}

export default function ScheduleCalendarEventWrapper({
  children,
  event,
  showScheduleLink,
}: ScheduleCalendarEventWrapperProps): React.JSX.Element {
  const classes = useStyles()
  const [anchorEl, setAnchorEl] = useState<HTMLButtonElement | null>(null)

  const [showEditDialog, setShowEditDialog] = useState('')
  const [showDeleteDialog, setShowDeleteDialog] = useState('')

  const { setOverrideDialog, onEditTempSched, onDeleteTempSched } = useContext(
    OverrideDialogContext,
  )
  const open = Boolean(anchorEl)
  const id = open ? 'shift-popover' : undefined

  function handleClick(e: MouseEvent<HTMLButtonElement>): void {
    setAnchorEl(e.currentTarget)
  }

  function handleCloseShiftInfo(): void {
    setAnchorEl(null)
  }

  function handleKeyDown(e: KeyboardEvent<HTMLButtonElement>): void {
    if (e.key === 'Enter' || e.key === ' ') {
      setAnchorEl(e.currentTarget)
    }
  }

  function handleShowOverrideForm(calEvent: OnCallShiftEvent): void {
    handleCloseShiftInfo()

    setOverrideDialog({
      variantOptions: ['replace', 'remove'],
      removeUserReadOnly: true,
      defaultValue: {
        start: event.start.toISOString(),
        end: event.end.toISOString(),
        removeUserID: calEvent.userID,
      },
    })
  }

  function renderTempSchedButtons(
    calEvent: TempSchedEvent | TempSchedShiftEvent,
  ): React.JSX.Element {
    // don't display actions on events in the past
    if (DateTime.fromJSDate(calEvent.end) <= DateTime.utc()) {
      return <React.Fragment />
    }

    return (
      <div className={classes.cardActionContainer}>
        <CardActions
          secondaryActions={[
            {
              icon: <EditIcon data-cy='edit-temp-sched' fontSize='small' />,
              label: 'Edit',
              handleOnClick: () => onEditTempSched(calEvent.tempSched),
            },
            {
              icon: <DeleteIcon data-cy='delete-temp-sched' fontSize='small' />,
              label: 'Delete',
              handleOnClick: () => onDeleteTempSched(calEvent.tempSched),
            },
          ]}
        />
      </div>
    )
  }

  function renderOverrideButtons(calEvent: OverrideEvent): React.JSX.Element {
    return (
      <div className={classes.cardActionContainer}>
        <CardActions
          secondaryActions={[
            {
              icon: <EditIcon fontSize='small' />,
              label: 'Edit',
              handleOnClick: () => {
                handleCloseShiftInfo()
                setShowEditDialog(calEvent.override.id)
              },
            },
            {
              icon: <DeleteIcon fontSize='small' />,
              label: 'Delete',
              handleOnClick: () => {
                handleCloseShiftInfo()
                setShowDeleteDialog(calEvent.override.id)
              },
            },
          ]}
        />
      </div>
    )
  }

  function renderShiftButtons(calEvent: OnCallShiftEvent): React.JSX.Element {
    return (
      <React.Fragment>
        <Grid item className={classes.flexGrow} />
        <Grid item>
          <Button
            data-cy='override'
            size='small'
            onClick={() => handleShowOverrideForm(calEvent)}
            variant='contained'
            title={`Temporarily remove ${calEvent.title} from this schedule`}
          >
            Override Shift
          </Button>
        </Grid>
      </React.Fragment>
    )
  }

  function renderButtons(): React.JSX.Element {
    if (showScheduleLink) {
      const id = event?.targetID ?? ''
      return (
        <React.Fragment>
          <Grid item className={classes.flexGrow} />
          <Grid item>
            <Button
              variant='contained'
              component={AppLink}
              to={/schedules/ + id}
            >
              Visit Schedule
            </Button>
          </Grid>
        </React.Fragment>
      )
    }
    if (DateTime.fromJSDate(event.end) <= DateTime.utc())
      return <React.Fragment />
    if (event.type === 'tempSched')
      return renderTempSchedButtons(event as TempSchedEvent)
    if (event.type === 'tempSchedShift')
      return renderTempSchedButtons(event as TempSchedShiftEvent)
    if (event.type === 'override')
      return renderOverrideButtons(event as OverrideEvent)

    return renderShiftButtons(event as OnCallShiftEvent)
  }

  function renderOverrideDescription(calEvent: OverrideEvent): React.JSX.Element {
    const getDesc = (addUser?: User, removeUser?: User): React.JSX.Element => {
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
          {getDesc(
            calEvent.override.addUser ?? undefined,
            calEvent.override.removeUser ?? undefined,
          )}
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
  function renderShiftInfo(): React.JSX.Element {
    const fmt = (date: Date): string =>
      DateTime.fromJSDate(date).toLocaleString(DateTime.DATETIME_FULL)

    return (
      <Grid container spacing={1}>
        <Grid item xs={12}>
          <Typography variant='body2'>
            <b>{showScheduleLink ? event?.user?.name : event?.targetName}</b>
          </Typography>
        </Grid>
        {event.type === 'override' &&
          renderOverrideDescription(event as OverrideEvent)}
        <Grid item xs={12}>
          <Typography variant='body2'>
            {`${fmt(event.start)}  –  ${fmt(event.end)}`}
          </Typography>
        </Grid>
        {renderButtons()}
      </Grid>
    )
  }

  return (
    <React.Fragment>
      <Popover
        id={id}
        open={open}
        anchorEl={anchorEl}
        onClose={handleCloseShiftInfo}
        anchorOrigin={{
          vertical: 'top',
          horizontal: 'left',
        }}
        transformOrigin={{
          vertical: 'top',
          horizontal: 'left',
        }}
        slotProps={{
          paper: {
            // @ts-expect-error - DOM attr for tests
            'data-cy': 'shift-tooltip',
          },
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
      <Suspense>
        {showEditDialog && (
          <ScheduleOverrideEditDialog
            overrideID={showEditDialog}
            onClose={() => setShowEditDialog('')}
          />
        )}
        {showDeleteDialog && (
          <ScheduleOverrideDeleteDialog
            overrideID={showDeleteDialog}
            onClose={() => setShowDeleteDialog('')}
          />
        )}
      </Suspense>
    </React.Fragment>
  )
}
