import React from 'react'
import { gql, useQuery } from '@apollo/client'
import { FormControlLabel, Switch } from '@material-ui/core'

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
  scheduleID: string
  label?: string | ((tz: string) => string)
}

export function ScheduleTZFilter(
  props: ScheduleTZFilterProps,
): JSX.Element | null {
  const [zone, setZone] = useURLParam('tz', 'local')
  const { data, loading, error } = useQuery(tzQuery, {
    pollInterval: 0,
    variables: { id: props.scheduleID },
  })
  const tz = data?.schedule?.timeZone

  let label = ''
  if (error) {
    label = 'Error: ' + (error.message || error)
  } else if (!data && loading) {
    label = 'Fetching timezone information...'
  } else {
    switch (typeof props.label) {
      case 'function':
        label = props.label(tz)
        break
      case 'string':
        label = props.label
        break
      default:
        label = `Show times in ${tz}`
        break
    }
  }

  return (
    <FormControlLabel
      control={
        <Switch
          checked={zone !== 'local'}
          onChange={(e) => setZone(e.target.checked ? tz : 'local')}
          value={tz}
          disabled={Boolean(loading || error)}
        />
      }
      label={label}
    />
  )
}
