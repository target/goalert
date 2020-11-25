import React, { useMemo } from 'react'
import { gql, useQuery } from '@apollo/client'
import { FormControlLabel, Switch } from '@material-ui/core'
import { DateTime } from 'luxon'

import { useURLParam } from '../actions/hooks'

const tzQuery = gql`
  query($id: ID!) {
    schedule(id: $id) {
      id
      timeZone
    }
  }
`

interface ScheduleTZFilterProps {
  label: (timeZone: string) => string
  scheduleID: string
}

export function ScheduleTZFilter(
  props: ScheduleTZFilterProps,
): JSX.Element | null {
  const [zone, setZone] = useURLParam<string>('tz', 'local')
  const localTZ = useMemo(() => DateTime.local().zoneName, [])
  const { data, loading, error } = useQuery(tzQuery, {
    pollInterval: 0,
    variables: { id: props.scheduleID },
  })

  const scheduleTZ = data?.schedule?.timeZone

  if (localTZ === scheduleTZ) {
    return null
  }

  let label = ''
  if (error) {
    label = 'Error: ' + (error.message || error)
  } else if (!data && loading) {
    label = 'Fetching timezone information...'
  } else {
    label = props.label
      ? props.label(scheduleTZ)
      : `Show times in ${scheduleTZ}`
  }

  return (
    <FormControlLabel
      control={
        <Switch
          checked={zone !== 'local'}
          onChange={(e) => setZone(e.target.checked ? scheduleTZ : 'local')}
          value={scheduleTZ}
          disabled={Boolean(loading || error)}
        />
      }
      label={label}
    />
  )
}
