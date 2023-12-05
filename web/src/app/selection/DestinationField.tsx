import React from 'react'
import { TextFieldProps } from '@mui/material/TextField'
import { DestinationType } from '../../schema'
import DestinationInputDirect from './DestinationInputDirect'
import { useDestinationType } from '../util/useDestinationTypes'

export default function DestinationField(
  props: TextFieldProps & { value: string; destType: DestinationType },
): React.ReactNode {
  const dest = useDestinationType(props.destType)

  return dest.requiredFields.map((field, idx) => {
    if (field.isSearchSelectable)
      throw new Error('query select not implemented')

    return (
      <DestinationInputDirect
        key={idx}
        {...props}
        config={field}
        disabled={props.disabled || !dest.enabled}
      />
    )
  })
}
