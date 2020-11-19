import React from 'react'
import { FormControlLabel, Switch } from '@material-ui/core'
import { useURLParam } from '../actions/hooks'

interface TimeZoneSwitchProps {
  option: string
}

function TimeZoneSwitch({ option }: TimeZoneSwitchProps): JSX.Element {
  const [zone, setZone] = useURLParam<string>('tz', 'local')

  return (
    <FormControlLabel
      control={
        <Switch
          checked={zone === option}
          onChange={(e) => setZone(e.target.checked ? option : 'local')}
          value={zone}
        />
      }
      label={`Configure in ${option}`}
    />
  )
}

export default TimeZoneSwitch
