import React from 'react'
import IconButton from '@material-ui/core/IconButton'
import makeStyles from '@material-ui/core/styles/makeStyles'
import Tooltip from '@material-ui/core/Tooltip/Tooltip'
import ScheduleIcon from '@material-ui/icons/Schedule'
import Delete from '@material-ui/icons/Delete'
import Error from '@material-ui/icons/Error'
import _ from 'lodash'
import { DateTime, Interval } from 'luxon'

import { Shift } from './sharedUtils'
import FlatList, {
  FlatListItem,
  FlatListListItem,
  FlatListNotice,
} from '../../lists/FlatList'
import { UserAvatar } from '../../util/avatars'
import { useUserInfo } from '../../util/useUserInfo'
import { parseInterval } from '../../util/shifts'
import { useScheduleTZ } from './hooks'
import { CircularProgress } from '@material-ui/core'
import { splitAtMidnight } from '../../util/luxon-helpers'
import {
  fmtTime,
  getCoverageGapItems,
  getSubheaderItems,
  getOutOfBoundsItems,
  Sortable,
  sortItems,
} from './shiftsListUtil'

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
  value,
  onRemove,
  scheduleID,
  handleCoverageGapClick,
}: TempSchedShiftsListProps): JSX.Element {
  const classes = useStyles()
  const { q, zone } = useScheduleTZ(scheduleID)
  let shifts = useUserInfo(value)

  // wait for zone
  if (q.loading || zone === '') {
    return (
      <div className={classes.spinContainer}>
        <CircularProgress />
      </div>
    )
  }

  if (edit) {
    shifts = shifts.filter(
      (s) => DateTime.fromISO(s.end, { zone }) > DateTime.now().setZone(zone),
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

    const subheaderItems = getSubheaderItems(schedInterval, shifts, zone)
    const coverageGapItems = getCoverageGapItems(
      schedInterval,
      shifts,
      zone,
      handleCoverageGapClick,
    )
    const outOfBoundsItems = getOutOfBoundsItems(schedInterval, shifts, zone)

    const shiftItems = (() => {
      return _.flatMap(shifts, (s) => {
        const shiftInv = parseInterval(s, zone)
        const isValid = schedInterval.engulfs(shiftInv)
        const dayInvs = splitAtMidnight(shiftInv)

        return dayInvs.map((inv, index) => {
          const startTime = fmtTime(inv.start)
          const endTime = fmtTime(inv.end)

          let subText = ''
          if (inv.length('hours') === 24) {
            // shift spans all day
            subText = 'All day'
          } else if (inv.engulfs(shiftInv)) {
            // shift is inside the day
            subText = `From ${startTime} to ${endTime}`
          } else if (inv.end === shiftInv.end) {
            subText = `Active until ${endTime}`
          } else {
            // shift starts and continues on for the rest of the day
            subText = `Active starting at ${startTime}\n`
          }

          return {
            scrollIntoView: true,
            id: DateTime.fromISO(s.start).toISO() + s.userID + index.toString(),
            title: s.user.name,
            subText,
            userID: s.userID,
            icon: <UserAvatar userID={s.userID} />,
            secondaryAction:
              index === 0 ? (
                <div className={classes.secondaryActionWrapper}>
                  {!isValid && (
                    <Tooltip
                      title='This shift extends beyond the start and/or end of this temporary schedule'
                      placement='left'
                    >
                      <Error color='error' />
                    </Tooltip>
                  )}
                  <IconButton
                    aria-label='delete shift'
                    onClick={() => onRemove(s)}
                  >
                    <Delete />
                  </IconButton>
                </div>
              ) : null,
            at: inv.start,
            itemType: 'shift',
          } as Sortable<FlatListItem>
        })
      })
    })()

    const startItem = (() => {
      let details = `Starts at ${fmtTime(DateTime.fromISO(start, { zone }))}`
      let message = ''

      if (
        edit &&
        DateTime.fromISO(start, { zone }) < DateTime.now().setZone(zone)
      ) {
        message = 'Currently active'
        details = 'Historical shifts will not be displayed'
      }

      return {
        id: 'sched-start_' + start,
        type: 'OK',
        icon: <ScheduleIcon />,
        message,
        details,
        at: DateTime.fromISO(start, { zone }),
        itemType: 'start',
      } as Sortable<FlatListNotice>
    })()

    const endItem = (() => {
      const at = DateTime.fromISO(end, { zone })
      const details = at.equals(at.startOf('day'))
        ? 'Ends at midnight'
        : 'Ends at ' + fmtTime(at)

      return {
        id: 'sched-end_' + end,
        type: 'OK',
        icon: <ScheduleIcon />,
        message: '',
        details,
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
