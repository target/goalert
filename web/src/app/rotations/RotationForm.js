import React, { useMemo, useState } from 'react'
import p from 'prop-types'
import {
  TextField,
  Grid,
  MenuItem,
  Typography,
  makeStyles,
  Switch,
  FormControlLabel,
} from '@material-ui/core'
import { startCase } from 'lodash'
import { DateTime } from 'luxon'
import { useQuery, gql } from '@apollo/client'

import { FormContainer, FormField } from '../forms'
import { TimeZoneSelect } from '../selection'
import { ISODateTimePicker } from '../util/ISOPickers'
import NumberField from '../util/NumberField'
import Spinner from '../loading/components/Spinner'

const query = gql`
  query upcomingHandoffTimes($input: UpcomingHandoffTimesInput) {
    upcomingHandoffTimes(input: $input)
  }
`

const rotationTypes = ['hourly', 'daily', 'weekly']

const useStyles = makeStyles({
  listNone: {
    listStyle: 'none',
  },
  bolder: {
    fontWeight: 'bolder',
  },
  flex: {
    display: 'flex',
  },
  height7rem: {
    height: '7rem',
  },
  noSpacing: {
    margin: 0,
    padding: 0,
  },
})

// getHours converts a count and one of ['hourly', 'daily', 'weekly']
// into length in hours e.g. (2, daily) => 48
function getHours(count, unit) {
  const lookup = {
    hourly: 1,
    daily: 24,
    weekly: 24 * 7,
  }
  return lookup[unit] * count
}

export default function RotationForm(props) {
  const { value } = props
  const classes = useStyles()
  const localZone = DateTime.local().zone.name
  const [configInZone, setConfigInZone] = useState(false)
  const configZone = configInZone ? value.timeZone : 'local'

  const [minStart, maxStart] = useMemo(() => [
    DateTime.local().minus({ year: 1 }),
    DateTime.local().plus({ year: 1 }),
  ])

  const { data, loading, error } = useQuery(query, {
    variables: {
      input: {
        start: value.start,
        timeZone: value.timeZone,
        hours: getHours(value.shiftLength, value.type),
        count: 3,
      },
    },
  })

  const isCalculating = !data || loading

  const upcomingHandoffs = isCalculating
    ? []
    : data.upcomingHandoffTimes.map((iso) =>
        DateTime.fromISO(iso)
          .setZone(configZone)
          .toLocaleString(DateTime.DATETIME_FULL),
      )

  return (
    <FormContainer optionalLabels {...props}>
      <Grid container spacing={2}>
        <Grid item xs={12}>
          <FormField
            fullWidth
            component={TextField}
            name='name'
            label='Name'
            required
          />
        </Grid>
        <Grid item xs={12}>
          <FormField
            fullWidth
            component={TextField}
            multiline
            name='description'
            label='Description'
            required
          />
        </Grid>
        <Grid item xs={6}>
          <FormField
            fullWidth
            component={TextField}
            select
            required
            label='Rotation Type'
            name='type'
          >
            {rotationTypes.map((type) => (
              <MenuItem value={type} key={type}>
                {startCase(type)}
              </MenuItem>
            ))}
          </FormField>
        </Grid>
        <Grid item xs={6}>
          <FormField
            fullWidth
            component={NumberField}
            required
            type='number'
            name='shiftLength'
            label='Shift Length'
            min={1}
            max={9000}
          />
        </Grid>
        <Grid item xs={localZone === value.timeZone ? 12 : 6}>
          <FormField
            fullWidth
            component={TimeZoneSelect}
            multiline
            name='timeZone'
            fieldName='timeZone'
            label='Time Zone'
            required
          />
        </Grid>
        {localZone !== value.timeZone && (
          <Grid item xs={6} className={classes.flex}>
            <Grid container justify='center'>
              <FormControlLabel
                control={
                  <Switch
                    checked={configInZone}
                    onChange={() => setConfigInZone(!configInZone)}
                    value={value.timeZone}
                  />
                }
                label={`Configure in ${value.timeZone}`}
              />
            </Grid>
          </Grid>
        )}

        <Grid item xs={12}>
          <FormField
            fullWidth
            component={ISODateTimePicker}
            timeZone={configZone}
            label='Initial Handoff Time'
            name='start'
            min={minStart.toISO()}
            max={maxStart.toISO()}
            required
          />
        </Grid>

        <Grid item xs={12} className={classes.height7rem}>
          <Typography variant='body2' className={classes.bolder}>
            Upcoming Handoff times:
          </Typography>
          <ol className={classes.noSpacing}>
            {upcomingHandoffs.map((text, i) => (
              <Typography
                key={i}
                component='li'
                className={classes.listNone}
                variant='body2'
              >
                {text}
              </Typography>
            ))}
          </ol>
          {isCalculating && <Spinner text='Calculating...' />}
          {error && (
            <Typography variant='body2' color='error'>
              {error.message}
            </Typography>
          )}
        </Grid>
      </Grid>
    </FormContainer>
  )
}

RotationForm.propTypes = {
  value: p.shape({
    name: p.string.isRequired,
    description: p.string.isRequired,
    timeZone: p.string.isRequired,
    type: p.oneOf(rotationTypes).isRequired,
    shiftLength: p.number.isRequired,
    start: p.string.isRequired,
  }).isRequired,

  errors: p.arrayOf(
    p.shape({
      field: p.oneOf([
        'name',
        'description',
        'timeZone',
        'type',
        'start',
        'shiftLength',
      ]).isRequired,
      message: p.string.isRequired,
    }),
  ),

  onChange: p.func.isRequired,
}
