import React, { ReactNode } from 'react'
import { TempSchedValue } from './sharedUtils'
import { Grid, Typography } from '@mui/material'
import { gql, useQuery } from 'urql'
import TempSchedShiftsList from './TempSchedShiftsList'

const query = gql`
  query schedShifts($id: ID!, $start: ISOTimestamp!, $end: ISOTimestamp!) {
    schedule(id: $id) {
      id
      shifts(start: $start, end: $end) {
        userID
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

interface TempSchedConfirmationProps {
  scheduleID: string
  value: TempSchedValue
}

export default function TempSchedConfirmation({
  scheduleID,
  value,
}: TempSchedConfirmationProps): ReactNode {
  const [{ data, fetching }] = useQuery({
    query,
    variables: {
      id: scheduleID,
      start: value.start,
      end: value.end,
    },
  })

  if (fetching) return 'Loading...'

  return (
    <Grid container spacing={4}>
      <Grid item xs={6}>
        <Typography variant='h6'>Current Temporary Schedule</Typography>
        <TempSchedShiftsList
          confirmationStep
          scheduleID={scheduleID}
          value={data.schedule.shifts}
          start={value.start}
          end={value.end}
          edit={false}
          compareRemovals={value.shifts}
        />
      </Grid>
      <Grid item xs={6}>
        <Typography variant='h6'>Proposed Changes</Typography>
        <TempSchedShiftsList
          confirmationStep
          scheduleID={scheduleID}
          value={value.shifts}
          start={value.start}
          end={value.end}
          edit={false}
          compareAdditions={data.schedule.shifts}
        />
      </Grid>
    </Grid>
  )
}
