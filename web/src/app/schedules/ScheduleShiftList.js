import React, { useMemo, useState } from 'react'
import { DateTime, Duration, Interval } from 'luxon'
import p from 'prop-types'
import FlatList from '../lists/FlatList'
import gql from 'graphql-tag'
import { relativeDate } from '../util/timeFormat'
import {
  Card,
  Grid,
  FormControlLabel,
  Switch,
  TextField,
  MenuItem,
  makeStyles,
} from '@material-ui/core'
import { UserAvatar } from '../util/avatars'
import PageActions from '../util/PageActions'
import FilterContainer from '../util/FilterContainer'
import { UserSelect } from '../selection'
import { useURLParam, useResetURLParams } from '../actions'
import { ScheduleTZFilter } from './ScheduleTZFilter'
import ScheduleNewOverrideFAB from './ScheduleNewOverrideFAB'
import ScheduleOverrideCreateDialog from './ScheduleOverrideCreateDialog'
import { ISODatePicker } from '../util/ISOPickers'
import { useQuery } from '@apollo/client'

// query name is important, as it's used for refetching data after mutations
const query = gql`
  query scheduleShifts($id: ID!, $start: ISOTimestamp!, $end: ISOTimestamp!) {
    schedule(id: $id) {
      id
      shifts(start: $start, end: $end) {
        user {
          id
          name
        }
        start
        end
        truncated
      }
    }
  }
`

const durString = (dur) => {
  if (dur.months) {
    return `${dur.months} month${dur.months > 1 ? 's' : ''}`
  }
  if (dur.days % 7 === 0) {
    const weeks = dur.days / 7
    return `${weeks} week${weeks > 1 ? 's' : ''}`
  }
  return `${dur.days} day${dur.days > 1 ? 's' : ''}`
}

const useStyles = makeStyles({
  datePicker: {
    width: '100%',
  },
})

function ScheduleShiftList({ scheduleID }) {
  const classes = useStyles()

  const [create, setCreate] = useState(null)
  const [specifyDuration, setSpecifyDuration] = useState(false)
  const [isClear, setIsClear] = useState(false)

  const [duration, setDuration] = useURLParam('duration', 'P14D')
  const [zone] = useURLParam('tz', 'local')
  const [userFilter, setUserFilter] = useURLParam('userFilter', [])
  const [activeOnly, setActiveOnly] = useURLParam('activeOnly', false)

  const defaultStart = useMemo(
    () => DateTime.fromObject({ zone }).startOf('day').toISO(),
    [zone],
  )
  const [_start, setStart] = useURLParam('start', defaultStart)
  const start = useMemo(() => (activeOnly ? DateTime.utc().toISO() : _start), [
    activeOnly,
    _start,
  ])

  const end = DateTime.fromISO(start, { zone })
    .plus(Duration.fromISO(duration))
    .toISO()

  const handleFilterReset = useResetURLParams(
    'userFilter',
    'start',
    'activeOnly',
    'tz',
    'duration',
  )

  const { data } = useQuery(query, {
    variables: {
      id: scheduleID,
      start,
      end,
    },
  })

  function items() {
    const _shifts =
      data?.schedule?.shifts?.map((s) => ({
        ...s,
        userID: s.user.id,
        userName: s.user.name,
      })) ?? []

    let shifts = _shifts
      .filter((s) => !userFilter.length || userFilter.includes(s.userID))
      .map((s) => ({
        ...s,
        start: DateTime.fromISO(s.start, { zone }),
        end: DateTime.fromISO(s.end, { zone }),
        interval: Interval.fromDateTimes(
          DateTime.fromISO(s.start, { zone }),
          DateTime.fromISO(s.end, { zone }),
        ),
      }))

    if (activeOnly) {
      const now = DateTime.fromObject({ zone })
      shifts = shifts.filter((s) => s.interval.contains(now))
    }

    if (!shifts.length) return []

    const displaySpan = Interval.fromDateTimes(
      DateTime.fromISO(start, { zone }).startOf('day'),
      DateTime.fromISO(end, { zone }).startOf('day'),
    )

    const result = []
    displaySpan.splitBy({ days: 1 }).forEach((day) => {
      const dayShifts = shifts.filter((s) => day.overlaps(s.interval))
      if (!dayShifts.length) return
      result.push({
        subHeader: relativeDate(day.start),
      })
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
          shiftDetails = `Active until${
            s.truncated ? ' at least' : ''
          } ${endTime}`
        } else {
          // shift starts and continues on for the rest of the day
          shiftDetails = `Active after ${startTime}`
        }
        result.push({
          title: s.userName,
          subText: shiftDetails,
          icon: <UserAvatar userID={s.userID} />,
        })
      })
    })

    return result
  }

  function renderDurationSelector() {
    // Dropdown options (in ISO_8601 format)
    // https://en.wikipedia.org/wiki/ISO_8601#Durations
    const quickOptions = ['P1D', 'P3D', 'P7D', 'P14D', 'P1M']
    const clamp = (min, max, value) => Math.min(max, Math.max(min, value))

    if (quickOptions.includes(duration) && !specifyDuration) {
      return (
        <TextField
          select
          fullWidth
          label='Time Limit'
          disabled={activeOnly}
          value={duration}
          onChange={(e) => {
            e.target.value === 'SPECIFY'
              ? setSpecifyDuration(true)
              : setDuration(e.target.value)
          }}
        >
          {quickOptions.map((opt) => (
            <MenuItem value={opt} key={opt}>
              {durString(Duration.fromISO(opt))}
            </MenuItem>
          ))}
          <MenuItem value='SPECIFY'>Specify...</MenuItem>
        </TextField>
      )
    }
    return (
      <TextField
        fullWidth
        label='Time Limit (days)'
        value={isClear ? '' : Duration.fromISO(duration).as('days')}
        disabled={activeOnly}
        max={30}
        min={1}
        type='number'
        onBlur={() => setIsClear(false)}
        onChange={(e) => {
          setIsClear(e.target.value === '')
          if (Number.isNaN(parseInt(e.target.value, 10))) {
            return
          }
          setDuration(
            Duration.fromObject({
              days: clamp(1, 30, parseInt(e.target.value, 10)),
            }).toISO(),
          )
        }}
      />
    )
  }

  const dur = Duration.fromISO(duration)
  const timeStr = durString(dur)

  const zoneText = zone === 'local' ? 'local time' : zone
  const userText = userFilter.length ? ' for selected users' : ''
  const note = activeOnly
    ? `Showing currently active shifts${userText} in ${zoneText}.`
    : `Showing shifts${userText} up to ${timeStr} from ${DateTime.fromISO(
        start,
        {
          zone,
        },
      ).toLocaleString()} in ${zoneText}.`
  return (
    <React.Fragment>
      <PageActions>
        <FilterContainer
          onReset={() => {
            handleFilterReset()
            setSpecifyDuration(false)
          }}
        >
          <Grid item xs={12}>
            <FormControlLabel
              control={
                <Switch
                  checked={activeOnly}
                  onChange={(e) => setActiveOnly(e.target.checked)}
                  value='activeOnly'
                />
              }
              label='Active shifts only'
            />
          </Grid>
          <Grid item xs={12}>
            <ScheduleTZFilter scheduleID={scheduleID} />
          </Grid>
          <Grid item xs={12}>
            <ISODatePicker
              className={classes.datePicker}
              disabled={activeOnly}
              label='Start Date'
              name='filterStart'
              value={start}
              onChange={(v) => setStart(v)}
            />
          </Grid>
          <Grid item xs={12}>
            {renderDurationSelector()}
          </Grid>
          <Grid item xs={12}>
            <UserSelect
              label='Filter users...'
              multiple
              value={userFilter}
              onChange={setUserFilter}
            />
          </Grid>
        </FilterContainer>
        <ScheduleNewOverrideFAB onClick={(variant) => setCreate(variant)} />
      </PageActions>
      <Card style={{ width: '100%' }}>
        <FlatList headerNote={note} items={items()} />
      </Card>
      {create && (
        <ScheduleOverrideCreateDialog
          scheduleID={scheduleID}
          variant={create}
          onClose={() => setCreate(null)}
        />
      )}
    </React.Fragment>
  )
}

ScheduleShiftList.propTypes = {
  scheduleID: p.string.isRequired,
}

export default ScheduleShiftList
