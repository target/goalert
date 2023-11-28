import React from 'react'
import { Grid, Typography } from '@mui/material'
import { DateTime } from 'luxon'
import { gql, useQuery } from 'urql'
import { RotationType, ISODuration, CreateRotationInput } from '../../schema'
import Spinner from '../loading/components/Spinner'
import { Time } from '../util/Time'

const query = gql`
  query calcRotationHandoffTimes($input: CalcRotationHandoffTimesInput) {
    calcRotationHandoffTimes(input: $input)
  }
`

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

interface RotationFormHandoffTimesProps {
  value: CreateRotationInput
}

export default function RotationFormHandoffTimes({
  value,
}: RotationFormHandoffTimesProps): JSX.Element {
  const [{ data, fetching, error }] = useQuery({
    query,
    variables: {
      input: {
        handoff: value.start,
        timeZone: value.timeZone,
        shiftLength: getShiftDuration(value.shiftLength as number, value.type),
        count: 3,
      },
    },
  })
  const isCalculating = !data || fetching

  const isHandoffValid = DateTime.fromISO(value.start).isValid
  const nextHandoffs = isCalculating ? [] : data.calcRotationHandoffTimes

  return (
    <Grid item xs={12} sx={{ height: '7rem' }}>
      <Typography variant='body2' sx={{ fontWeight: 'bolder' }}>
        Upcoming Handoff times:
      </Typography>
      {isHandoffValid ? (
        <ol style={{ margin: 0, padding: 0 }}>
          {nextHandoffs.map((time: string, i: number) => (
            <Typography
              key={i}
              component='li'
              variant='body2'
              sx={{ listStyle: 'none' }}
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
  )
}
