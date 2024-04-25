import { gql, useQuery } from '@apollo/client'
import { GroupAdd } from '@mui/icons-material'
import {
  Button,
  Card,
  FormControlLabel,
  Grid,
  Switch,
  TextField,
  MenuItem,
  Tooltip,
} from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import { DateTime, DateTimeFormatOptions, Duration, Interval } from 'luxon'
import React, { Suspense, useCallback, useMemo, useState } from 'react'
import CreateFAB from '../lists/CreateFAB'
import FlatList, { FlatListListItem } from '../lists/FlatList'
import { UserSelect } from '../selection'
import { UserAvatar } from '../util/avatars'
import FilterContainer from '../util/FilterContainer'
import { useURLParam, useResetURLParams } from '../actions'
import { ISODatePicker } from '../util/ISOPickers'
import { useScheduleTZ } from './useScheduleTZ'
import { relativeDate } from '../util/timeFormat'
import { useIsWidthDown } from '../util/useWidth'
import { OverrideDialog, OverrideDialogContext } from './ScheduleDetails'
import ScheduleOverrideDialog from './ScheduleOverrideDialog'
import {
  Shift,
  TempSchedValue,
  defaultTempSchedValue,
} from './temp-sched/sharedUtils'
import TempSchedDialog from './temp-sched/TempSchedDialog'

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

const durString = (dur: Duration): string => {
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

interface ScheduleShiftListProps {
  scheduleID: string
}

function ScheduleShiftList({
  scheduleID,
}: ScheduleShiftListProps): JSX.Element {
  const classes = useStyles()
  const isMobile = useIsWidthDown('md')

  const [specifyDuration, setSpecifyDuration] = useState(false)
  const [isClear, setIsClear] = useState(false)

  const [overrideDialog, setOverrideDialog] = useState<OverrideDialog | null>(
    null,
  )
  const { zone: scheduleTZ } = useScheduleTZ(scheduleID)
  const [configTempSchedule, setConfigTempSchedule] =
    useState<TempSchedValue | null>(null)
  const onNewTempSched = useCallback(
    () => setConfigTempSchedule(defaultTempSchedValue(scheduleTZ)),
    [],
  )
  const [duration, setDuration] = useURLParam<string>('duration', 'P14D')
  const [userFilter, setUserFilter] = useURLParam<string[]>('userFilter', [])
  const [activeOnly, setActiveOnly] = useURLParam<boolean>('activeOnly', false)

  const defaultStart = DateTime.local().startOf('day').toISO()
  const [_start, setStart] = useURLParam('start', defaultStart)
  const start = useMemo(
    () => (activeOnly ? DateTime.utc().toISO() : _start),
    [activeOnly, _start],
  )

  const end = DateTime.fromISO(start).plus(Duration.fromISO(duration)).toISO()

  const dur = Duration.fromISO(duration)
  const timeStr = durString(dur)

  const tzAbbr = DateTime.local({ zone: scheduleTZ }).toFormat('ZZZZ')
  const localTzAbbr = DateTime.local({ zone: 'local' }).toFormat('ZZZZ')

  const userText = userFilter.length ? ' for selected users' : ''
  const note = activeOnly
    ? `Showing currently active shifts${userText} in ${localTzAbbr}.`
    : `Showing shifts${userText} up to ${timeStr} from ${DateTime.fromISO(
        start,
      ).toLocaleString()} in ${localTzAbbr}.`

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

  // getShiftDetails uses the inputted shift (s) and day interval (day) values and returns either a string or a hover tooltip element.
  // Defaults to the user's local tz, with the schedule tz info offered as a tooltip
  function getShiftDetails(
    s: {
      start: DateTime
      end: DateTime
      interval: Interval
      truncated: boolean
    },
    day: Interval,
  ): JSX.Element | string {
    const locale: DateTimeFormatOptions = {
      hour: 'numeric',
      minute: 'numeric',
    }
    let schedLocale: DateTimeFormatOptions = locale
    const _localStartTime = DateTime.fromISO(s.start.toISO(), {
      zone: 'local',
    })
    const _localEndTime = DateTime.fromISO(s.end.toISO(), {
      zone: 'local',
    })

    // show full date in tooltip if the schedule tz day is not the locale day
    if (s.start.day > _localStartTime.day || s.end.day > _localEndTime.day) {
      schedLocale = {
        hour: 'numeric',
        minute: 'numeric',
        month: 'short',
        day: '2-digit',
      }
    }

    const schedStartTime = s.start.toLocaleString(schedLocale)
    const schedEndTime = s.end.toLocaleString(schedLocale)
    const localStartTime = _localStartTime.toLocaleString(locale)
    const localEndTime = _localEndTime.toLocaleString(locale)

    if (s.interval.engulfs(day)) {
      return 'All day'
    }

    if (day.engulfs(s.interval)) {
      const scheduleTZDetails = `From ${schedStartTime} to ${schedEndTime} ${tzAbbr}`
      const localTZDetails = `From ${localStartTime} to ${localEndTime} ${localTzAbbr}`
      return (
        <Tooltip
          title={<span data-testid='shift-tooltip'>{scheduleTZDetails}</span>}
          placement='right'
        >
          <span data-testid='shift-details'>{localTZDetails}</span>
        </Tooltip>
      )
    }

    if (day.contains(s.end)) {
      const scheduleTZDetails = `Active until${
        s.truncated ? ' at least' : ''
      } ${schedEndTime} ${tzAbbr}`
      const localTZDetails = `Active until${
        s.truncated ? ' at least' : ''
      } ${localEndTime} ${localTzAbbr}`
      return (
        <Tooltip
          title={<span data-testid='shift-tooltip'>{scheduleTZDetails}</span>}
          placement='right'
        >
          <span data-testid='shift-details'>{localTZDetails}</span>
        </Tooltip>
      )
    }

    // shift starts and continues on for the rest of the day
    const scheduleTZDetails = `Active after ${schedStartTime} ${tzAbbr}`
    const localTZDetails = `Active after ${localStartTime} ${localTzAbbr}`
    return (
      <Tooltip
        title={<span data-testid='shift-tooltip'>{scheduleTZDetails}</span>}
        placement='right'
      >
        <span data-testid='shift-details'>{localTZDetails}</span>
      </Tooltip>
    )
  }

  function items(): FlatListListItem[] {
    const _shifts: Shift[] =
      data?.schedule?.shifts?.map((s: Shift) => ({
        ...s,
        userID: s.user?.id || '',
      })) ?? []

    let shifts = _shifts
      .filter((s) => !userFilter.length || userFilter.includes(s.userID))
      .map((s) => ({
        ...s,
        start: DateTime.fromISO(s.start, { zone: scheduleTZ }),
        end: DateTime.fromISO(s.end, { zone: scheduleTZ }),
        interval: Interval.fromDateTimes(
          DateTime.fromISO(s.start),
          DateTime.fromISO(s.end),
        ),
      }))
    if (activeOnly) {
      const now = DateTime.local()
      shifts = shifts.filter((s) => s.interval.contains(now))
    }

    if (!shifts.length) return []

    const displaySpan = Interval.fromDateTimes(
      DateTime.fromISO(start).startOf('day'),
      DateTime.fromISO(end).startOf('day'),
    )

    const result: FlatListListItem[] = []
    displaySpan.splitBy({ days: 1 }).forEach((day) => {
      const dayShifts = shifts.filter((s) => day.overlaps(s.interval))
      if (!dayShifts.length) return
      result.push({
        subHeader: relativeDate(day.start),
      })
      dayShifts.forEach((s) => {
        result.push({
          title: s.user?.name || '',
          subText: getShiftDetails(s, day),
          icon: <UserAvatar userID={s.userID} />,
        })
      })
    })

    return result
  }

  function renderDurationSelector(): JSX.Element {
    // Dropdown options (in ISO_8601 format)
    // https://en.wikipedia.org/wiki/ISO_8601#Durations
    const quickOptions = ['P1D', 'P3D', 'P7D', 'P14D', 'P1M']
    const clamp = (min: number, max: number, value: number): number =>
      Math.min(max, Math.max(min, value))

    if (quickOptions.includes(duration) && !specifyDuration) {
      return (
        <TextField
          select
          fullWidth
          label='Time Limit'
          disabled={activeOnly}
          value={duration}
          onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => {
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
        InputProps={{ inputProps: { min: 1, max: 30 } }}
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

  return (
    <OverrideDialogContext.Provider
      value={{
        onNewTempSched,
        setOverrideDialog,
        onEditTempSched: () => {},
        onDeleteTempSched: () => {},
      }}
    >
      <Card style={{ width: '100%' }}>
        <FlatList
          headerNote={note}
          items={items()}
          headerAction={
            <React.Fragment>
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
                  <ISODatePicker
                    className={classes.datePicker}
                    disabled={activeOnly}
                    label='Start Date'
                    name='filterStart'
                    value={start}
                    onChange={(v: string) => setStart(v)}
                    fullWidth
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
              {!isMobile && (
                <Button
                  variant='contained'
                  startIcon={<GroupAdd />}
                  onClick={() =>
                    setOverrideDialog({
                      variantOptions: ['replace', 'remove', 'add', 'temp'],
                      removeUserReadOnly: false,
                    })
                  }
                  sx={{ ml: 1 }}
                >
                  Create Override
                </Button>
              )}
            </React.Fragment>
          }
        />
      </Card>
      {isMobile && (
        <CreateFAB
          title='Create Override'
          onClick={() =>
            setOverrideDialog({
              variantOptions: ['replace', 'remove', 'add', 'temp'],
              removeUserReadOnly: false,
            })
          }
        />
      )}

      {/* create dialogs */}
      <Suspense>
        {overrideDialog && (
          <ScheduleOverrideDialog
            defaultValue={overrideDialog.defaultValue}
            variantOptions={overrideDialog.variantOptions}
            scheduleID={scheduleID}
            onClose={() => setOverrideDialog(null)}
            removeUserReadOnly={overrideDialog.removeUserReadOnly}
          />
        )}
        {configTempSchedule && (
          <TempSchedDialog
            value={configTempSchedule}
            onClose={() => setConfigTempSchedule(null)}
            scheduleID={scheduleID}
          />
        )}
      </Suspense>
    </OverrideDialogContext.Provider>
  )
}

export default ScheduleShiftList
