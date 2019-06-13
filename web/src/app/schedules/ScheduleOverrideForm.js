import React from 'react'
import p from 'prop-types'
import { FormContainer, FormField } from '../forms'
import {
  Grid,
  InputAdornment,
  IconButton,
  Typography,
  withStyles,
} from '@material-ui/core'
import { ScheduleTZFilter } from './ScheduleTZFilter'
import { connect } from 'react-redux'
import { urlParamSelector } from '../selectors'
import { DateRange, ChevronRight, ChevronLeft } from '@material-ui/icons'
import { DateTimePicker } from 'material-ui-pickers'
import { DateTime } from 'luxon'
import { UserSelect } from '../selection'
import Query from '../util/Query'
import gql from 'graphql-tag'
import { mapOverrideUserError } from './util'
import DialogContentError from '../dialogs/components/DialogContentError'

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

const styles = theme => ({
  tzNote: {
    display: 'flex',
    alignItems: 'center',
  },
})

@connect(state => ({ zone: urlParamSelector(state)('tz', 'local') }))
@withStyles(styles)
export default class ScheduleOverrideForm extends React.PureComponent {
  static propTypes = {
    scheduleID: p.string.isRequired,

    value: p.shape({
      addUserID: p.string.isRequired,
      removeUserID: p.string.isRequired,
      start: p.string.isRequired,
      end: p.string.isRequired,
    }).isRequired,

    add: p.bool,
    remove: p.bool,

    errors: p.arrayOf(
      p.shape({
        field: p.oneOf(['addUserID', 'removeUserID', 'userID', 'start', 'end'])
          .isRequired,
        message: p.string.isRequired,
      }),
    ),

    onChange: p.func.isRequired,
  }

  render() {
    const userError = this.props.errors.find(e => e.field === 'userID')
    return (
      <Query
        query={query}
        variables={{ id: userError ? userError.details.CONFLICTING_ID : '' }}
        noPoll
        skip={!userError}
        noSpin
        render={({ data }) => this.renderForm(data)}
      />
    )
  }

  renderForm(data) {
    const { add, remove, zone, errors, value, ...formProps } = this.props
    const userError = errors.find(e => e.field === 'userID')
    const formErrors = errors
      .filter(e => e.field !== 'userID')
      .concat(
        userError ? mapOverrideUserError(data.userOverride, value, zone) : [],
      )

    return (
      <FormContainer
        optionalLabels
        errors={formErrors}
        value={value}
        {...formProps}
      >
        <Grid container spacing={2}>
          <Grid
            item
            xs={12}
            sm={12}
            md={6}
            className={this.props.classes.tzNote}
          >
            <Typography
              // variant='caption'
              color='textSecondary'
              style={{ fontStyle: 'italic' }}
            >
              Start and end time shown in{' '}
              {zone === 'local' ? 'local time' : zone}.
            </Typography>
          </Grid>
          <Grid item xs={12} sm={12} md={6}>
            {/* Purposefully leaving out of form, as it's only used for converting display times. */}
            <ScheduleTZFilter
              label={tz => `Configure in ${tz}`}
              scheduleID={this.props.scheduleID}
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
              fullWidth
              component={DateTimePicker}
              mapValue={value => DateTime.fromISO(value, { zone })}
              mapOnChangeValue={value => value.toISO()}
              showTodayButton
              required
              name='start'
              leftArrowIcon={<ChevronLeft />}
              rightArrowIcon={<ChevronRight />}
              InputProps={{
                endAdornment: (
                  <InputAdornment position='end'>
                    <IconButton>
                      <DateRange />
                    </IconButton>
                  </InputAdornment>
                ),
              }}
            />
          </Grid>
          <Grid item xs={12}>
            <FormField
              fullWidth
              component={DateTimePicker}
              mapValue={value => DateTime.fromISO(value, { zone })}
              mapOnChangeValue={value => value.toISO()}
              showTodayButton
              name='end'
              required
              leftArrowIcon={<ChevronLeft />}
              rightArrowIcon={<ChevronRight />}
              InputProps={{
                endAdornment: (
                  <InputAdornment position='end'>
                    <IconButton>
                      <DateRange />
                    </IconButton>
                  </InputAdornment>
                ),
              }}
            />
          </Grid>
          {userError && <DialogContentError error={userError.message} />}
        </Grid>
      </FormContainer>
    )
  }
}
