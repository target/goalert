import React from 'react'
import { DateTime, Duration, Interval } from 'luxon'
import p from 'prop-types'
import FlatList from '../lists/FlatList'
import gql from 'graphql-tag'
import { withQuery } from '../util/Query'
import { connect } from 'react-redux'
import { urlParamSelector } from '../selectors'
import { relativeDate } from '../util/timeFormat'
import {
  Card,
  Grid,
  FormControlLabel,
  Switch,
  TextField,
  MenuItem,
  withStyles,
} from '@material-ui/core'
import { UserAvatar } from '../util/avatars'
import PageActions from '../util/PageActions'
import FilterContainer from '../util/FilterContainer'
import { UserSelect } from '../selection'
import { setURLParam, resetURLParams } from '../actions'
import { ScheduleTZFilter } from './ScheduleTZFilter'
import ScheduleNewOverrideFAB from './ScheduleNewOverrideFAB'
import ScheduleOverrideCreateDialog from './ScheduleOverrideCreateDialog'
import { ISODatePicker } from '../util/ISOPickers'

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

const mapQueryToProps = ({ data }) => {
  return {
    shifts: data.schedule.shifts.map((s) => ({
      ...s,
      userID: s.user.id,
      userName: s.user.name,
    })),
  }
}
const mapPropsToQueryProps = ({ scheduleID, start, end }) => ({
  variables: {
    id: scheduleID,
    start,
    end,
  },
})

const mapStateToProps = (state) => {
  const duration = urlParamSelector(state)('duration', 'P14D')
  const zone = urlParamSelector(state)('tz', 'local')
  let start = urlParamSelector(state)(
    'start',
    DateTime.fromObject({ zone }).startOf('day').toISO(),
  )

  const activeOnly = urlParamSelector(state)('activeOnly', false)
  if (activeOnly) {
    start = DateTime.fromObject({ zone }).toISO()
  }

  const end = DateTime.fromISO(start, { zone })
    .plus(Duration.fromISO(duration))
    .toISO()

  return {
    start,
    end,
    userFilter: urlParamSelector(state)('userFilter', []),
    activeOnly,
    duration,
    zone,
  }
}

const mapDispatchToProps = (dispatch) => {
  return {
    handleUserFilterSelect: (value) =>
      dispatch(setURLParam('userFilter', value)),
    handleActiveOnlySwitch: (value) =>
      dispatch(setURLParam('activeOnly', value)),
    handleSetDuration: (value) =>
      dispatch(setURLParam('duration', value, 'P14D')),
    handleSetStart: (value) => dispatch(setURLParam('start', value)),
    handleFilterReset: () =>
      dispatch(
        resetURLParams('userFilter', 'start', 'activeOnly', 'tz', 'duration'),
      ),
  }
}

const styles = {
  datePicker: {
    width: '100%',
  },
}

@withStyles(styles)
@connect(mapStateToProps, mapDispatchToProps)
@withQuery(query, mapQueryToProps, mapPropsToQueryProps)
export default class ScheduleShiftList extends React.PureComponent {
  static propTypes = {
    scheduleID: p.string.isRequired,

    // provided by connect
    start: p.string.isRequired,
    end: p.string.isRequired,
    zone: p.string.isRequired,

    // provided by withQuery
    shifts: p.arrayOf(
      p.shape({
        start: p.string.isRequired,
        end: p.string.isRequired,
        userID: p.string.isRequired,
        userName: p.string.isRequired,
        truncated: p.bool,
      }),
    ),
  }

  static defaultProps = {
    shifts: [],
  }

  state = {
    create: null,
    specifyDuration: false,
    isClear: false,
  }

  items() {
    const {
      shifts: _shifts,
      start,
      end,
      userFilter,
      activeOnly,
      zone,
    } = this.props

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

  renderDurationSelector() {
    // Dropdown options (in ISO_8601 format)
    // https://en.wikipedia.org/wiki/ISO_8601#Durations
    const quickOptions = ['P1D', 'P3D', 'P7D', 'P14D', 'P1M']
    const clamp = (min, max, value) => Math.min(max, Math.max(min, value))

    if (
      quickOptions.includes(this.props.duration) &&
      !this.state.specifyDuration
    ) {
      return (
        <TextField
          select
          fullWidth
          label='Time Limit'
          disabled={this.props.activeOnly}
          value={this.props.duration}
          onChange={(e) => {
            e.target.value === 'SPECIFY'
              ? this.setState({ specifyDuration: true })
              : this.props.handleSetDuration(e.target.value)
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
        value={
          this.state.isClear
            ? ''
            : Duration.fromISO(this.props.duration).as('days')
        }
        disabled={this.props.activeOnly}
        max={30}
        min={1}
        type='number'
        onBlur={() => {
          this.setState({ isClear: false })
        }}
        onChange={(e) => {
          this.setState({ isClear: e.target.value === '' })
          let val = e.target.value
          if (parseInt(val, 10) > 30) {
            val = '30'
          } else if (parseInt(val, 10) < 1) {
            val = '1'
          }
          if (parseInt(e.target.value) > 0) {
            this.props.handleSetDuration(
              Duration.fromObject({
                days: clamp(1, 30, parseInt(e.target.value, 10)),
              }).toISO(),
            )
          }
        }}
      />
    )
  }

  render() {
    const zone = this.props.zone
    const dur = Duration.fromISO(this.props.duration)

    const timeStr = durString(dur)

    const zoneText = zone === 'local' ? 'local time' : zone
    const userText = this.props.userFilter.length ? ' for selected users' : ''
    const note = this.props.activeOnly
      ? `Showing currently active shifts${userText} in ${zoneText}.`
      : `Showing shifts${userText} up to ${timeStr} from ${DateTime.fromISO(
          this.props.start,
          {
            zone,
          },
        ).toLocaleString()} in ${zoneText}.`
    return (
      <React.Fragment>
        <PageActions>
          <FilterContainer
            onReset={() => {
              this.props.handleFilterReset()
              this.setState({ specifyDuration: false })
            }}
          >
            <Grid item xs={12}>
              <FormControlLabel
                control={
                  <Switch
                    checked={this.props.activeOnly}
                    onChange={(e) =>
                      this.props.handleActiveOnlySwitch(e.target.checked)
                    }
                    value='activeOnly'
                  />
                }
                label='Active shifts only'
              />
            </Grid>
            <Grid item xs={12}>
              <ScheduleTZFilter scheduleID={this.props.scheduleID} />
            </Grid>
            <Grid item xs={12}>
              <ISODatePicker
                className={this.props.classes.datePicker}
                disabled={this.props.activeOnly}
                label='Start Date'
                name='filterStart'
                value={this.props.start}
                onChange={(v) => this.props.handleSetStart(v)}
              />
            </Grid>
            <Grid item xs={12}>
              {this.renderDurationSelector()}
            </Grid>
            <Grid item xs={12}>
              <UserSelect
                label='Filter users...'
                multiple
                value={this.props.userFilter}
                onChange={this.props.handleUserFilterSelect}
              />
            </Grid>
          </FilterContainer>
          <ScheduleNewOverrideFAB
            onClick={(variant) => this.setState({ create: variant })}
          />
        </PageActions>
        <Card style={{ width: '100%' }}>
          <FlatList headerNote={note} items={this.items()} />
        </Card>
        {this.state.create && (
          <ScheduleOverrideCreateDialog
            scheduleID={this.props.scheduleID}
            variant={this.state.create}
            onClose={() => this.setState({ create: null })}
          />
        )}
      </React.Fragment>
    )
  }
}
