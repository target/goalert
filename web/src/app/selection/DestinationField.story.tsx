import React from 'react'
import DestinationField, { DestinationFieldProps } from './DestinationField'

export function DestinationFieldValueWrapper(
  props: DestinationFieldProps,
): React.ReactNode {
  const [value, setValue] = React.useState(props.value)

  return (
    <DestinationField
      value={value}
      onChange={(v) => {
        setValue(v)
      }}
      destType={props.destType}
      disabled={props.disabled}
    />
  )
}
