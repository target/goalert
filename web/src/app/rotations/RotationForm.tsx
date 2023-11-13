import React from 'react'
import {
  TextField,
  Grid,
  MenuItem,
  Typography,
  FormHelperText,
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
import { CreateRotationInput, ISODuration, RotationType } from '../../schema'
import { Time } from '../util/Time'

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

const rotationTypes = ['hourly', 'daily', 'weekly', 'monthly']

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

// getShiftDuration converts a count and one of ['hourly', 'daily', 'weekly', 'monthly']
// into the shift length to ISODuration.
function getShiftDuration(count: number, type: RotationType): ISODuration {
  switch (type) {
    case 'monthly':
      return `P${count}M`
    case 'weekly':
      return `P${count}W`
    case 'daily':
      return `P${count}D`
    case 'hourly':
      return `PT${count}H`
    default:
      throw new Error('unknown rotation type: ' + type)
  }
}

const sameAsLocal = (t: string, z: string): boolean => {
  const inZone = DateTime.fromISO(t, { zone: z })
  const inLocal = DateTime.fromISO(t, { zone: 'local' })
  return inZone.toFormat('Z') === inLocal.toFormat('Z')
}

export default function RotationForm(props: RotationFormProps): React.ReactNode {
  const { value } = props
  const classes = useStyles()

  const { data, loading, error } = useQuery(query, {
    variables: {
      input: {
        handoff: value.start,
        timeZone: value.timeZone,
        shiftLength: getShiftDuration(value.shiftLength as number, value.type),
        count: 3,
      },
    },
  })

  const isCalculating = !data || loading

  const isHandoffValid = DateTime.fromISO(value.start).isValid
  const handoffWarning =
    DateTime.fromISO(value.start).day > 28 && value.type === 'monthly'
  const nextHandoffs = isCalculating ? [] : data.calcRotationHandoffTimes

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
        <Grid item xs={12}>
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

        <Grid item xs={12}>
          <FormField
            fullWidth
            component={ISODateTimePicker}
            timeZone={value.timeZone}
            label={`Handoff Time (${value.timeZone})`}
            name='start'
            required
            hint={
              <React.Fragment>
                {sameAsLocal(value.start, value.timeZone) ? undefined : (
                  <Time time={value.start} suffix=' local time' />
                )}
                {handoffWarning && (
                  <FormHelperText
                    sx={{
                      color: (theme) => theme.palette.warning.main,
                      marginLeft: 0,
                      marginRight: 0,
                    }}
                    data-cy='handoff-warning'
                  >
                    Unintended handoff behavior may occur when date starts after
                    the 28th
                  </FormHelperText>
                )}
              </React.Fragment>
            }
          />
        </Grid>

        <Grid item xs={12} className={classes.handoffsContainer}>
          <Typography variant='body2' className={classes.handoffsTitle}>
            Upcoming Handoff times:
          </Typography>
          {isHandoffValid ? (
            <ol className={classes.handoffsList}>
              {nextHandoffs.map((time: string, i: number) => (
                <Typography
                  key={i}
                  component='li'
                  className={classes.handoffTimestamp}
                  variant='body2'
                >
                  <Time time={time} zone={value.timeZone} />
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
