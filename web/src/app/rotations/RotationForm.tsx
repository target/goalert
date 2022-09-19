import React, { useState } from 'react'
import {
  TextField,
  Grid,
  MenuItem,
  Typography,
  Switch,
  FormControlLabel,
} from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import { startCase } from 'lodash'
import { DateTime } from 'luxon'
import { useQuery, gql } from '@apollo/client'

import { FormContainer, FormField } from '../forms'
import { TimeZoneSelect } from '../selection'
import { ISODateTimePicker } from '../util/ISOPickers'
import NumberField from '../util/NumberField'
import Spinner from '../loading/components/Spinner'
import { FieldError } from '../util/errutil'
import { RotationType, CreateRotationInput } from '../../schema'

interface RotationFormProps {
  value: CreateRotationInput
  errors: FieldError[]
  onChange: (value: CreateRotationInput) => void
  disabled?: boolean
}

const query = gql`
  query calcRotationHandoffTimes($input: CalcRotationHandoffTimesInput) {
    calcRotationHandoffTimes(input: $input)
  }
`

const rotationTypes = ['hourly', 'daily', 'weekly']

const useStyles = makeStyles({
  handoffTimestamp: {
    listStyle: 'none',
  },
  handoffsTitle: {
    fontWeight: 'bolder',
  },
  tzContainer: {
    display: 'flex',
  },
  handoffsContainer: {
    height: '7rem',
  },
  handoffsList: {
    margin: 0,
    padding: 0,
  },
})

// getHours converts a count and one of ['hourly', 'daily', 'weekly']
// into length in hours e.g. (2, daily) => 48
function getHours(count: number, unit: RotationType): number {
  const lookup = {
    hourly: 1,
    daily: 24,
    weekly: 24 * 7,
  }
  return lookup[unit] * count
}

export default function RotationForm(props: RotationFormProps): JSX.Element {
  const { value } = props
  const classes = useStyles()
  const localZone = DateTime.local().zone.name
  const [configInZone, setConfigInZone] = useState(false)
  const configZone = configInZone ? value.timeZone : 'local'

  const { data, loading, error } = useQuery(query, {
    variables: {
      input: {
        handoff: value.start,
        from: value.start,
        timeZone: value.timeZone,
        shiftLengthHours: getHours(value.shiftLength as number, value.type),
        count: 3,
      },
    },
  })

  const isCalculating = !data || loading

  const isHandoffValid = DateTime.fromISO(value.start).isValid
  const nextHandoffs = isCalculating
    ? []
    : data.calcRotationHandoffTimes.map((iso: string) =>
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
          <Grid item xs={6} className={classes.tzContainer}>
            <Grid container justifyContent='center'>
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
            label='Handoff Time'
            name='start'
            required
          />
        </Grid>

        <Grid item xs={12} className={classes.handoffsContainer}>
          <Typography variant='body2' className={classes.handoffsTitle}>
            Upcoming Handoff times:
          </Typography>
          {isHandoffValid ? (
            <ol className={classes.handoffsList}>
              {nextHandoffs.map((text: string, i: number) => (
                <Typography
                  key={i}
                  component='li'
                  className={classes.handoffTimestamp}
                  variant='body2'
                >
                  {text}
                </Typography>
              ))}
            </ol>
          ) : (
            <Typography variant='body2' color='textSecondary'>
              Please enter a valid handoff time.
            </Typography>
          )}

          {isCalculating && isHandoffValid && <Spinner text='Calculating...' />}
          {error && isHandoffValid && (
            <Typography variant='body2' color='error'>
              {error.message}
            </Typography>
          )}
        </Grid>
      </Grid>
    </FormContainer>
  )
}
