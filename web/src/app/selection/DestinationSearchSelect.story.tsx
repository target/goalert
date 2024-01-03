import React from 'react'
import DestinationSearchSelect, {
  DestinationSearchSelectProps,
} from './DestinationSearchSelect'

export function DestinationSearchSelectWrapper(
  props: DestinationSearchSelectProps,
): React.ReactNode {
  const [value, setValue] = React.useState(props.value)

  return (
    <DestinationSearchSelect
      value={value}
      onChange={(v) => {
        setValue(v.target.value)
      }}
      config={props.config}
      destType={props.destType}
      disabled={props.disabled}
    />
  )
}
