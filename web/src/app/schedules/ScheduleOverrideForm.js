import React from 'react'
import p from 'prop-types'
import { FormContainer, FormField } from '../forms'
import {
  Grid,
  InputAdornment,
  IconButton,
  Typography,
  makeStyles,
} from '@material-ui/core'
import { useQuery } from 'react-apollo'
import { ScheduleTZFilter } from './ScheduleTZFilter'
import { useSelector } from 'react-redux'
import { urlParamSelector } from '../selectors'
import { DateRange } from '@material-ui/icons'
import { KeyboardDateTimePicker } from '@material-ui/pickers'
import { DateTime } from 'luxon'
import { UserSelect } from '../selection'
import gql from 'graphql-tag'
import { mapOverrideUserError } from './util'
import DialogContentError from '../dialogs/components/DialogContentError'
import _ from 'lodash-es'

const query = gql`
  query($id: ID!) {
    userOverride(id: $id) {
      id
      addUser {
        id
        name
      }
      removeUser {
        id
        name
      }
      start
      end
    }
  }
`

const useStyles = makeStyles({
  tzNote: {
    display: 'flex',
    alignItems: 'center',
  },
})

export default function ScheduleOverrideForm(props) {
  const { onChange, value } = props
  const { add, remove, errors = [], scheduleID, ...formProps } = props

  const classes = useStyles()
  const params = useSelector(urlParamSelector)
  const zone = params('tz', 'local')

  const conflictingUserFieldError = props.errors.find(
    e => e && e.field === 'userID',
  )

  // used to grab conflicting errors from pre-existing overrides
  const { data } = useQuery(query, {
    variables: {
      id: _.get(conflictingUserFieldError, 'details.CONFLICTING_ID', ''),
    },
    pollInterval: 0,
    skip: !conflictingUserFieldError,
  })

  const userConflictErrors = errors
    .filter(e => e.field !== 'userID')
    .concat(
      conflictingUserFieldError
        ? mapOverrideUserError(_.get(data, 'userOverride'), value, zone)
        : [],
    )

  return (
    <FormContainer
      optionalLabels
      errors={errors.concat(userConflictErrors)}
      {...formProps}
      value={value}
    >
      <Grid container spacing={2}>
        <Grid item xs={12} sm={12} md={6} className={classes.tzNote}>
          <Typography
            // variant='caption'
            color='textSecondary'
            style={{ fontStyle: 'italic' }}
          >
            Start and end time shown in {zone === 'local' ? 'local time' : zone}
            .
          </Typography>
        </Grid>
        <Grid item xs={12} sm={12} md={6}>
          {/* Purposefully leaving out of form, as it's only used for converting display times. */}
          <ScheduleTZFilter
            label={tz => `Configure in ${tz}`}
            scheduleID={scheduleID}
            onChange={newZone => {
              // update value in form container with new TZ
              onChange({
                ...value,
                start: value.start.setZone(newZone),
                end: value.end.setZone(newZone),
              })
            }}
          />
        </Grid>
        {remove && (
          <Grid item xs={12}>
            <FormField
              fullWidth
              component={UserSelect}
              name='removeUserID'
              label={add && remove ? 'User to be Replaced' : 'User to Remove'}
              required
            />
          </Grid>
        )}
        {add && (
          <Grid item xs={12}>
            <FormField
              fullWidth
              component={UserSelect}
              required
              name='addUserID'
              label='User to Add'
            />
          </Grid>
        )}
        <Grid item xs={12}>
          <FormField
            data-cy='start-date'
            fullWidth
            component={KeyboardDateTimePicker}
            showTodayButton
            required
            name='start'
            InputProps={{
              endAdornment: (
                <InputAdornment position='end'>
                  <IconButton>
                    <DateRange />
                  </IconButton>
                </InputAdornment>
              ),
            }}
            // keyboard picker props
            format='MM/dd/yyyy, hh:mm a'
            mask='__/__/____, __:__ _M'
            placeholder={DateTime.local()
              .startOf('day')
              .toFormat('MM/dd/yyyy, hh:mm a')}
            invalidDateMessage={null}
            validate={() => {
              if (!props.value.start.isValid) {
                return new Error('Invalid time')
              }
            }}
          />
        </Grid>
        <Grid item xs={12}>
          <FormField
            data-cy='end-date'
            fullWidth
            component={KeyboardDateTimePicker}
            showTodayButton
            name='end'
            required
            InputProps={{
              endAdornment: (
                <InputAdornment position='end'>
                  <IconButton>
                    <DateRange />
                  </IconButton>
                </InputAdornment>
              ),
            }}
            // keyboard picker props
            format='MM/dd/yyyy, hh:mm a'
            mask='__/__/____, __:__ _M'
            placeholder={DateTime.local()
              .endOf('day')
              .toFormat('MM/dd/yyyy, hh:mm a')}
            invalidDateMessage={null}
            validate={() => {
              if (!props.value.end.isValid) {
                return new Error('Invalid time')
              }
            }}
          />
        </Grid>
        {conflictingUserFieldError && (
          <DialogContentError error={conflictingUserFieldError.message} />
        )}
      </Grid>
    </FormContainer>
  )
}

ScheduleOverrideForm.propTypes = {
  scheduleID: p.string.isRequired,

  value: p.shape({
    addUserID: p.string.isRequired,
    removeUserID: p.string.isRequired,
    start: p.instanceOf(DateTime),
    end: p.instanceOf(DateTime),
  }).isRequired,

  add: p.bool,
  remove: p.bool,

  disabled: p.bool.isRequired,
  errors: p.arrayOf(
    p.shape({
      field: p.oneOf(['addUserID', 'removeUserID', 'userID', 'start', 'end'])
        .isRequired,
      message: p.string.isRequired,
    }),
  ),

  onChange: p.func.isRequired,
}
