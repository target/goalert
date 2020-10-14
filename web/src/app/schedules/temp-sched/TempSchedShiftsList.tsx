import React from 'react'
import { IconButton, makeStyles, Typography } from '@material-ui/core'
import FlatList from '../../lists/FlatList'
import { Shift } from './sharedUtils'
import { UserAvatar } from '../../util/avatars'
import { Delete } from '@material-ui/icons'
import { Warning } from '../../icons/components/Icons'
import { useUserInfo } from '../../util/useUserInfo'
import { DateTime, Interval } from 'luxon'
import { useURLParam } from '../../actions'
import { relativeDate } from '../../util/timeFormat'
import _ from 'lodash-es'
import { isBefore, isAfter } from '../../util/luxon-helpers'

const useStyles = makeStyles({
  secondaryActionWrapper: {
    display: 'flex',
    alignItems: 'center',
  },
})

type TempSchedShiftsListProps = {
  value: Shift[]
  onRemove: (shift: Shift) => void

  start: string
  end: string
}

type FlatListSub = {
  subHeader: string
}
type FlatListItem = {
  title: string
  subText: string
  icon?: JSX.Element
  secondaryAction?: JSX.Element | null
}

type FlatListListItem = FlatListSub | FlatListItem

export default function TempSchedShiftsList({
  start,
  end,
  value,
  onRemove,
}: TempSchedShiftsListProps): JSX.Element {
  const classes = useStyles()
  const _shifts = useUserInfo(value)
  const [zone] = useURLParam('tz', 'local')

  function items(): FlatListListItem[] {
    const shifts = _.sortBy(_shifts, 'start').map((s) => ({
      shift: s,
      added: false,
      start: DateTime.fromISO(s.start, { zone }),
      end: DateTime.fromISO(s.end, { zone }),
      interval: Interval.fromDateTimes(
        DateTime.fromISO(s.start, { zone }),
        DateTime.fromISO(s.end, { zone }),
      ),
      isValid: !isAfter(start, s.start) && !isBefore(end, s.end),
    }))

    if (!shifts.length) return []

    const scheduleDawn = DateTime.fromISO(start, { zone }).startOf('day')
    const scheduleDusk = DateTime.fromISO(end, { zone }).endOf('day')

    const firstShiftDawn = shifts[0].start.startOf('day')
    const lastShiftDusk = shifts[shifts.length - 1].end.endOf('day')

    const displaySpan = Interval.fromDateTimes(
      DateTime.min(scheduleDawn, firstShiftDawn),
      DateTime.max(scheduleDusk, lastShiftDusk),
    )

    const result: FlatListListItem[] = []
    displaySpan.splitBy({ days: 1 }).forEach((day) => {
      const dayShifts = shifts.filter((s) => day.overlaps(s.interval))
      if (!dayShifts.length) return

      // render subheader for each day
      result.push({
        subHeader: relativeDate(day.start),
      })

      // craft user friendly shift string
      dayShifts.forEach((s) => {
        let shiftDetails = ''
        const startTime = s.start.toLocaleString({
          hour: 'numeric',
          minute: 'numeric',
        })
        const endTime = s.end.toLocaleString({
          hour: 'numeric',
          minute: 'numeric',
        })

        if (s.interval.engulfs(day)) {
          // shift (s.interval) spans all day
          shiftDetails = 'All day'
        } else if (day.engulfs(s.interval)) {
          // shift is inside the day
          shiftDetails = `From ${startTime} to ${endTime}`
        } else if (day.contains(s.end)) {
          shiftDetails = `Active until ${endTime}`
        } else {
          // shift starts and continues on for the rest of the day
          shiftDetails = `Active starting at ${startTime}`
        }

        result.push({
          title: s.shift.user.name,
          subText: shiftDetails,
          icon: <UserAvatar userID={s.shift.userID} />,
          secondaryAction: s.added ? null : (
            <div className={classes.secondaryActionWrapper}>
              {!s.isValid && (
                <Warning
                  message='This shift extends beyond the start and/or end of this temporary schedule'
                  placement='left'
                />
              )}
              <IconButton onClick={() => onRemove(s.shift)}>
                <Delete />
              </IconButton>
            </div>
          ),
        })

        s.added = true
      })
    })

    return result
  }

  return (
    <React.Fragment>
      <Typography variant='subtitle1' component='h3'>
        Shifts
      </Typography>
      <FlatList
        items={items()}
        emptyMessage='Add a user to the left to get started.'
        dense
      />
    </React.Fragment>
  )
}
