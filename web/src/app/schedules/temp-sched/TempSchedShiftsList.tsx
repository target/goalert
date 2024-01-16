import React, { useEffect, useState } from 'react'
import IconButton from '@mui/material/IconButton'
import makeStyles from '@mui/styles/makeStyles'
import Tooltip from '@mui/material/Tooltip'
import { Shift } from './sharedUtils'
import ScheduleIcon from '@mui/icons-material/Schedule'
import Delete from '@mui/icons-material/Delete'
import Error from '@mui/icons-material/Error'
import _ from 'lodash'
import { DateTime, Duration, Interval } from 'luxon'

import FlatList, {
  FlatListItem,
  FlatListListItem,
  FlatListNotice,
} from '../../lists/FlatList'
import { UserAvatar } from '../../util/avatars'
import { useUserInfo } from '../../util/useUserInfo'
import { parseInterval } from '../../util/shifts'
import { useScheduleTZ } from '../useScheduleTZ'
import { CircularProgress } from '@mui/material'
import { splitShift } from '../../util/luxon-helpers'
import {
  getCoverageGapItems,
  getSubheaderItems,
  getOutOfBoundsItems,
  Sortable,
  sortItems,
} from './shiftsListUtil'
import { fmtLocal, fmtTime } from '../../util/timeFormat'

const useStyles = makeStyles({
  secondaryActionWrapper: {
    display: 'flex',
    alignItems: 'center',
  },
  spinContainer: {
    display: 'flex',
    alignItems: 'center',
    flexDirection: 'column',
    marginTop: '15rem',
  },
})

type TempSchedShiftsListProps = {
  value: Shift[]
  onRemove: (shift: Shift) => void
  shiftDur: Duration
  start: string
  end: string
  edit?: boolean
  scheduleID: string
  handleCoverageGapClick: (coverageGap: Interval) => void
}

export default function TempSchedShiftsList({
  edit,
  start,
  end,
  shiftDur,
  value,
  onRemove,
  scheduleID,
  handleCoverageGapClick,
}: TempSchedShiftsListProps): JSX.Element {
  const classes = useStyles()
  const { zone, isLocalZone } = useScheduleTZ(scheduleID)
  const [now, setNow] = useState(DateTime.now().setZone(zone))
  const shifts = useUserInfo(value)
  const [existingShifts] = useState(shifts)

  useEffect(() => {
    if (edit) {
      const interval = setTimeout(
        () => {
          setNow(DateTime.now().setZone(zone))
        },
        Duration.fromObject({ minutes: 1 }).as('millisecond'),
      )
      return () => clearTimeout(interval)
    }
  }, [now])

  // wait for zone
  if (zone === '') {
    return (
      <div className={classes.spinContainer}>
        <CircularProgress />
      </div>
    )
  }

  const schedInterval = parseInterval({ start, end }, zone)

  function items(): FlatListListItem[] {
    // render helpful message if interval is invalid
    // shouldn't ever be seen because of our validation checks, but just in case
    if (!schedInterval.isValid) {
      return [
        {
          id: 'invalid-sched-interval',
          type: 'ERROR',
          message: 'Invalid Start/End',
          details:
            'Oops! There was a problem with the interval selected for your temporary schedule. Please try again.',
        },
      ]
    }

    const coverageGapItems = getCoverageGapItems(
      schedInterval,
      shiftDur,
      shifts,
      zone,
      handleCoverageGapClick,
    )
    const subheaderItems = getSubheaderItems(
      schedInterval,
      shifts,
      shiftDur,
      zone,
    )

    const outOfBoundsItems = getOutOfBoundsItems(schedInterval, shifts, zone)

    const shiftItems = (() => {
      return _.flatMap(shifts, (s, idx) => {
        const shiftInv = parseInterval(s, zone)
        const isValid = schedInterval.engulfs(shiftInv)
        const fixedShifts = splitShift(shiftInv)

        return fixedShifts.map((inv, index) => {
          const startTime = fmtTime(
            s.displayStart ? s.displayStart : inv.start,
            zone,
            false,
            false,
          )
          const endTime = fmtTime(inv.end, zone, false, false)
          const shiftExists = existingShifts.find((shift) => {
            return (
              DateTime.fromISO(s.start).equals(DateTime.fromISO(shift.start)) &&
              DateTime.fromISO(s.end).equals(DateTime.fromISO(shift.end)) &&
              s.userID === shift.userID
            )
          })
          const isHistoricShift =
            Boolean(shiftExists?.userID) &&
            DateTime.fromISO(s.end, { zone }) < now

          let subText = ''
          let titleText = ''
          if (inv.length('hours') === 24) {
            // shift spans all day
            subText = 'All day'
          } else if (inv.engulfs(shiftInv)) {
            // shift is inside the day
            subText = `From ${startTime} to ${endTime}`
            titleText = `From ${fmtLocal(inv.start.toISO())} to ${fmtLocal(
              inv.end.toISO(),
            )}`
          } else if (inv.end === shiftInv.end) {
            subText = `Active until ${endTime}`
            titleText = `Active until ${fmtLocal(inv.end.toISO())}`
          } else {
            // shift starts and continues on for the rest of the day
            subText = `Active starting at ${startTime}\n`
            titleText = `Active starting at ${fmtLocal(inv.start.toISO())}`
          }

          return {
            scrollIntoView: true,
            id: DateTime.fromISO(s.start).toISO() + s.userID + index.toString(),
            title: s.user.name,
            subText: (
              <Tooltip title={!isLocalZone ? titleText : ''} placement='right'>
                <span>{subText}</span>
              </Tooltip>
            ),
            icon: <UserAvatar userID={s.userID} />,
            disabled: isHistoricShift,
            secondaryAction: index === 0 && (
              <div className={classes.secondaryActionWrapper}>
                {!isValid && !isHistoricShift && (
                  <Tooltip
                    title='This shift extends beyond the start and/or end of this temporary schedule'
                    placement='left'
                  >
                    <Error color='error' />
                  </Tooltip>
                )}
                {!isHistoricShift && (
                  <IconButton
                    data-cy={'delete shift index: ' + idx}
                    aria-label='delete shift'
                    onClick={() => onRemove(s)}
                  >
                    <Delete />
                  </IconButton>
                )}
              </div>
            ),
            at: inv.start,
            itemType: 'shift',
          } as Sortable<FlatListItem>
        })
      })
    })()

    const startItem = (() => {
      const active = edit && DateTime.fromISO(start, { zone }) < now

      const { message, details, at, itemType, tooltipTitle } = active
        ? {
            message: 'Currently active',
            details: 'Historical shifts are not editable',
            at: DateTime.min(
              DateTime.fromISO(start, { zone }),
              ...shifts.map((s) => DateTime.fromISO(s.start, { zone })),
            ).startOf('day'),
            itemType: 'active',
            tooltipTitle: '',
          }
        : {
            message: '',
            details: `Starts at ${fmtTime(start, zone, false, false)}`,
            at: DateTime.fromISO(start, { zone }),
            itemType: 'start',
            tooltipTitle: `Starts at ${fmtLocal(start)}`,
          }

      return {
        id: 'sched-start_' + start,
        type: 'OK',
        icon: <ScheduleIcon />,
        message,
        at,
        itemType,
        details: (
          <Tooltip title={!isLocalZone ? tooltipTitle : ''} placement='right'>
            <div>{details}</div>
          </Tooltip>
        ),
      } as Sortable<FlatListNotice>
    })()

    const endItem = (() => {
      const at = DateTime.fromISO(end, { zone })
      const details = at.equals(at.startOf('day'))
        ? 'Ends at midnight'
        : 'Ends at ' + fmtTime(at, zone, false, false)
      const detailsTooltip = `Ends at ${fmtLocal(end)}`

      return {
        id: 'sched-end_' + end,
        type: 'OK',
        icon: <ScheduleIcon />,
        message: '',
        details: (
          <Tooltip title={!isLocalZone ? detailsTooltip : ''} placement='right'>
            <div>{details}</div>
          </Tooltip>
        ),
        at,
        itemType: 'end',
      } as Sortable<FlatListNotice>
    })()

    return sortItems([
      ...shiftItems,
      ...coverageGapItems,
      ...subheaderItems,
      ...outOfBoundsItems,
      startItem,
      endItem,
    ])
  }

  return (
    <FlatList
      data-cy='shifts-list'
      items={items()}
      emptyMessage='Add a user to the left to get started.'
      dense
      transition
    />
  )
}
