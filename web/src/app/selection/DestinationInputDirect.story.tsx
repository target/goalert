import React from 'react'
import DestinationInputDirect, {
  DestinationInputDirectProps,
} from './DestinationInputDirect'

export function DestinationInputDirectValueWrapper(
  props: DestinationInputDirectProps,
): React.ReactNode {
  const [value, setValue] = React.useState(props.value)

  return (
    <DestinationInputDirect
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
