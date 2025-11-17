import React, { useEffect, useMemo, useState } from 'react'
import {
  Avatar,
  Button,
  Checkbox,
  Chip,
  FormControlLabel,
  Grid,
  Popover,
  IconButton,
  ButtonGroup,
  useTheme,
} from '@mui/material'
import Typography from '@mui/material/Typography'
import ToggleIcon from '@mui/icons-material/CompareArrows'
import _ from 'lodash'
import {
  dtToDuration,
  Shift,
  sortUsersByLastPickOrder,
  TempSchedValue,
} from './sharedUtils'
import { FormContainer, FormField } from '../../forms'
import { DateTime, Duration, Interval } from 'luxon'
import { FieldError } from '../../util/errutil'
import { isISOAfter } from '../../util/shifts'
import { useScheduleTZ } from '../useScheduleTZ'
import { ISODateTimePicker } from '../../util/ISOPickers'
import { UserSelect } from '../../selection'
import ClickableText from '../../util/ClickableText'
import NumberField from '../../util/NumberField'
import { fmtLocal } from '../../util/timeFormat'
import { User } from 'web/src/schema'
import {
  ArrowRight,
  ShuffleVariant,
  Sort,
  SortAlphabeticalAscending,
  SortAlphabeticalDescending,
} from 'mdi-material-ui'
import Chance from 'chance'
import { gql, useQuery } from 'urql'
const c = new Chance()

const query = gql`
  query ($id: ID!) {
    schedule(id: $id) {
      lastTempSchedPickOrder
    }
  }
`

type AddShiftsStepProps = {
  value: TempSchedValue
  onChange: (newValue: Shift[]) => void

  scheduleID: string
  associatedUsers: Array<User>
  showForm: boolean
  setShowForm: (showForm: boolean) => void
  shift: Shift
  setShift: (shift: Shift) => void
  isCustomShiftTimeRange: boolean
  setIsCustomShiftTimeRange: (bool: boolean) => void
  pickOrder: string[]
  setPickOrder: (pickOrder: string[]) => void
}

type SortType = 'A-Z' | 'Z-A' | 'RAND' | 'LAST-PICKS'

type DTShift = {
  userID: string
  span: Interval
}

function shiftsToDT(shifts: Shift[]): DTShift[] {
  return shifts.map((s) => ({
    userID: s.userID,
    span: Interval.fromDateTimes(
      DateTime.fromISO(s.start),
      DateTime.fromISO(s.end),
    ),
  }))
}

function DTToShifts(shifts: DTShift[]): Shift[] {
  return shifts.map((s) => ({
    userID: s.userID,
    start: s.span.start.toISO(),
    end: s.span.end.toISO(),
    truncated: false,
  }))
}

// mergeShifts will take the incoming shifts and merge them with
// the shifts stored in value. Using Luxon's Interval, overlaps
// and edge cases when merging are handled for us.
function mergeShifts(_shifts: Shift[]): Shift[] {
  const byUser = _.groupBy(shiftsToDT(_shifts), 'userID')

  return DTToShifts(
    _.flatten(
      _.values(
        _.mapValues(byUser, (shifts, userID) => {
          return Interval.merge(_.map(shifts, 'span')).map((span) => ({
            userID,
            span,
          }))
        }),
      ),
    ),
  )
}

export default function TempSchedAddNewShift({
  scheduleID,
  associatedUsers,
  onChange,
  value,
  shift,
  setShift,
  isCustomShiftTimeRange,
  setIsCustomShiftTimeRange,
  pickOrder,
  setPickOrder,
}: AddShiftsStepProps): JSX.Element {
  const theme = useTheme()
  const [submitted, setSubmitted] = useState(false)

  const [manualEntry, setManualEntry] = useState(true)
  const { q, zone, isLocalZone } = useScheduleTZ(scheduleID)

  const [sortType, setSortType] = useState<SortType>('A-Z')
  const [sortTypeAnchor, setSortTypeAnchor] =
    useState<HTMLButtonElement | null>(null)
  const sortPopoverOpen = Boolean(sortTypeAnchor)
  const sortTypeID = sortPopoverOpen ? 'sort-type-select' : undefined

  const [{ fetching, error, data }] = useQuery({
    query,
    variables: {
      id: scheduleID,
    },
  })
  const lastTempSchedPickOrder: Array<string> =
    data.schedule.lastTempSchedPickOrder

  const handleFilterTypeClick = (
    event: React.MouseEvent<HTMLButtonElement>,
  ): void => {
    setSortTypeAnchor(event.currentTarget)
  }

  const handleFilterTypeClose = (): void => {
    setSortTypeAnchor(null)
  }

  // set start equal to the temporary schedule's start
  // can't this do on mount since the step renderer puts everyone on the DOM at once
  useEffect(() => {
    if (zone === '') return

    setShift({
      start: value.start,
      end: DateTime.fromISO(value.start, { zone })
        .plus(value.shiftDur as Duration)
        .toISO(),
      userID: '',
      truncated: false,
    })
  }, [value.start, zone, value.shiftDur])

  // fieldErrors handles errors manually through the client
  // as this step form is nested inside the greater form
  // that makes the network request.
  function fieldErrors(s = submitted): FieldError[] {
    const result: FieldError[] = []
    const requiredMsg = 'this field is required'
    const add = (field: string, message: string): void => {
      result.push({ field, message } as FieldError)
    }

    if (!shift) return result
    if (s) {
      if (!shift.userID) add('userID', requiredMsg)
      if (!shift.start) add('start', requiredMsg)
      if (!shift.end) add('end', requiredMsg)
    }

    if (!isISOAfter(shift.end, shift.start)) {
      add('end', 'must be after shift start time')
      add('start', 'must be before shift end time')
    }

    return result
  }

  function handleAddShift(): void {
    if (fieldErrors(true).length) {
      setSubmitted(true)
      return
    }
    if (!shift) return // ts sanity check

    if (!pickOrder.includes(shift.userID)) {
      setPickOrder([...pickOrder, shift.userID])
    }

    onChange(mergeShifts(value.shifts.concat(shift)))
    const end = DateTime.fromISO(shift.end, { zone })
    setShift({
      userID: '',
      truncated: false,
      start: shift.end,
      end: end.plus(value.shiftDur as Duration).toISO(),
    })
    setIsCustomShiftTimeRange(false)
    setSubmitted(false)
  }

  function getUserIDCountInValue(userID: string): number {
    const uIDs = value.shifts.map((s) => s.userID)
    const count = uIDs.filter((id) => id === userID).length
    return count
  }

  function getChipColor(
    userID: string,
    count: number,
  ): 'primary' | 'success' | 'default' {
    if (shift.userID === userID) return 'primary'
    if (count > 0) return 'success'
    return 'default'
  }

  function handleSetSortType(sortType: SortType): void {
    setSortType(sortType)
    setSortTypeAnchor(null)
  }

  function sortFn(_a: User, _b: User): number {
    const a = _a.name
    const b = _b.name

    if (sortType === 'A-Z') {
      if (a > b) return 1
      if (a < b) return -1
      return 0
    }

    if (sortType === 'Z-A') {
      if (a < b) return 1
      if (a > b) return -1
      return 0
    }

    if (sortType === 'RAND') {
      return c.pickone([1, -1, 0])
    }

    if (sortType === 'LAST-PICKS') {
      return sortUsersByLastPickOrder(_a, _b, lastTempSchedPickOrder)
    }

    return 0
  }

  const users = useMemo(() => associatedUsers.sort(sortFn), [sortType])

  return (
    <FormContainer
      errors={fieldErrors()}
      value={shift}
      onChange={(val: Shift) => setShift(val)}
    >
      <Grid container spacing={2}>
        <Grid item xs={12} sx={{ display: 'flex', alignItems: 'center' }}>
          <Typography>Add Shift</Typography>
          <IconButton
            aria-describedby={sortTypeID}
            onClick={handleFilterTypeClick}
            color='primary'
            size='small'
            sx={{ ml: 0.5 }}
          >
            {sortType === 'A-Z' && (
              <SortAlphabeticalAscending fontSize='small' />
            )}
            {sortType === 'Z-A' && (
              <SortAlphabeticalDescending fontSize='small' />
            )}
            {sortType === 'RAND' && <ShuffleVariant fontSize='small' />}
            {sortType === 'LAST-PICKS' && !fetching && !error && (
              <Sort fontSize='small' />
            )}
          </IconButton>
          <Popover
            id={sortTypeID}
            open={sortPopoverOpen}
            anchorEl={sortTypeAnchor}
            onClose={handleFilterTypeClose}
            anchorOrigin={{
              vertical: 'bottom',
              horizontal: 'left',
            }}
          >
            <ButtonGroup orientation='vertical' variant='text'>
              <Button
                key='a-z'
                onClick={() => handleSetSortType('A-Z')}
                startIcon={<SortAlphabeticalAscending fontSize='small' />}
                sx={{ justifyContent: 'start', pl: 2, pr: 2 }}
              >
                Sort A-Z
              </Button>
              <Button
                key='z-a'
                onClick={() => handleSetSortType('Z-A')}
                startIcon={<SortAlphabeticalDescending fontSize='small' />}
                sx={{ justifyContent: 'start', pl: 2, pr: 2 }}
              >
                Sort Z-A
              </Button>
              <Button
                key='rand'
                onClick={() => handleSetSortType('RAND')}
                startIcon={<ShuffleVariant fontSize='small' />}
                sx={{ justifyContent: 'start', pl: 2, pr: 2 }}
              >
                Shuffle
              </Button>
              <Button
                key='last-picks'
                onClick={() => handleSetSortType('LAST-PICKS')}
                startIcon={<Sort fontSize='small' />}
                sx={{ justifyContent: 'start', pl: 2, pr: 2 }}
              >
                Sort By Last Pick Order
              </Button>
            </ButtonGroup>
          </Popover>
        </Grid>
        <Grid item xs={12} sx={{ mt: '-16px' }}>
          <Typography variant='caption' color='textSecondary'>
            Showing all users assigned to this schedule. Select a user to add to
            the next shift.
          </Typography>
        </Grid>

        <Grid item xs={12} container>
          {users.map((u) => {
            const count = getUserIDCountInValue(u.id)

            return (
              <Chip
                key={u.id}
                label={u.name}
                sx={{ m: 0.5 }}
                color={getChipColor(u.id, count)}
                onClick={() => {
                  setShift({
                    ...shift,
                    userID: u.id,
                  })
                }}
                icon={
                  count > 0 ? (
                    <Avatar
                      sx={{
                        width: 22,
                        height: 22,
                        fontSize: 15,
                        bgcolor:
                          theme.palette.mode === 'dark'
                            ? theme.palette.success.dark
                            : theme.palette.success.light,
                      }}
                    >
                      {count}
                    </Avatar>
                  ) : undefined
                }
              />
            )
          })}
        </Grid>

        <Grid item xs={12}>
          <FormField
            fullWidth
            component={UserSelect}
            label='Search for a user...'
            name='userID'
          />
        </Grid>
        <Grid item xs={6}>
          <FormControlLabel
            control={
              <Checkbox
                checked={!isCustomShiftTimeRange}
                data-cy='toggle-custom-off'
              />
            }
            label={
              <Typography
                color={!isCustomShiftTimeRange ? 'default' : 'textSecondary'}
                sx={{ fontStyle: 'italic' }}
              >
                Add user to next shift
              </Typography>
            }
            onChange={() => setIsCustomShiftTimeRange(false)}
          />
        </Grid>
        <Grid item xs={6}>
          <FormControlLabel
            control={
              <Checkbox
                checked={isCustomShiftTimeRange}
                data-cy='toggle-custom'
              />
            }
            label={
              <Typography
                color={isCustomShiftTimeRange ? 'default' : 'textSecondary'}
                sx={{ fontStyle: 'italic' }}
              >
                Add user to custom time range
              </Typography>
            }
            onChange={() => setIsCustomShiftTimeRange(true)}
          />
        </Grid>
        <Grid item xs={6}>
          <FormField
            fullWidth
            component={ISODateTimePicker}
            label='Shift Start'
            name='shift-start'
            fieldName='start'
            min={value.start}
            max={DateTime.fromISO(value.end, { zone })
              .plus({ year: 1 })
              .toISO()}
            mapOnChangeValue={(value: string, formValue: TempSchedValue) => {
              if (!manualEntry) {
                const diff = DateTime.fromISO(value, { zone }).diff(
                  DateTime.fromISO(formValue.start, { zone }),
                )
                formValue.end = DateTime.fromISO(formValue.end, { zone })
                  .plus(diff)
                  .toISO()
              }
              return value
            }}
            timeZone={zone}
            disabled={q.loading || !isCustomShiftTimeRange}
            hint={isLocalZone ? '' : fmtLocal(value?.start)}
          />
        </Grid>
        <Grid item xs={6}>
          {manualEntry ? (
            <FormField
              fullWidth
              component={ISODateTimePicker}
              label='Shift End'
              name='shift-end'
              fieldName='end'
              min={value.start}
              max={DateTime.fromISO(value.end, { zone })
                .plus({ year: 1 })
                .toISO()}
              hint={
                isCustomShiftTimeRange ? (
                  <React.Fragment>
                    {!isLocalZone && fmtLocal(value?.end)}
                    <div>
                      <ClickableText
                        data-cy='toggle-duration-on'
                        onClick={() => setManualEntry(false)}
                        endIcon={<ToggleIcon />}
                      >
                        Configure as duration
                      </ClickableText>
                    </div>
                  </React.Fragment>
                ) : null
              }
              timeZone={zone}
              disabled={q.loading || !isCustomShiftTimeRange}
            />
          ) : (
            <FormField
              fullWidth
              component={NumberField}
              label='Shift Duration (hours)'
              name='shift-end'
              fieldName='end'
              float
              // value held in form input
              mapValue={(nextVal: string, formValue: TempSchedValue) => {
                const nextValDT = DateTime.fromISO(nextVal, { zone })
                const formValDT = DateTime.fromISO(formValue?.start ?? '', {
                  zone,
                })
                const duration = dtToDuration(formValDT, nextValDT)
                return duration === -1 ? '' : duration.toString()
              }}
              // value held in state
              mapOnChangeValue={(
                nextVal: string,
                formValue: TempSchedValue,
              ) => {
                if (!nextVal) return ''
                return DateTime.fromISO(formValue.start, { zone })
                  .plus({ hours: parseFloat(nextVal) })
                  .toISO()
              }}
              step='any'
              min={0}
              disabled={q.loading || !isCustomShiftTimeRange}
              hint={
                isCustomShiftTimeRange ? (
                  <ClickableText
                    data-cy='toggle-duration-off'
                    onClick={() => setManualEntry(true)}
                    endIcon={<ToggleIcon />}
                  >
                    Configure as date/time
                  </ClickableText>
                ) : null
              }
            />
          )}
        </Grid>
        <Grid item xs={12} container justifyContent='flex-end'>
          <Button
            fullWidth
            data-cy='add-shift'
            color='secondary'
            variant='contained'
            onClick={handleAddShift}
            endIcon={<ArrowRight />}
          >
            Add Next Shift
          </Button>
        </Grid>
      </Grid>
    </FormContainer>
  )
}
