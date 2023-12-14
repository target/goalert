import React from 'react'
import { DestinationType, FieldValueInput } from '../../schema'
import DestinationInputDirect from './DestinationInputDirect'
import { useDestinationType } from '../util/useDestinationTypes'

export type DestinationFieldProps = {
  value: FieldValueInput[]
  onChange?: (value: FieldValueInput[]) => void
  destType: DestinationType

  disabled?: boolean
}

export default function DestinationField(
  props: DestinationFieldProps,
): React.ReactNode {
  const dest = useDestinationType(props.destType)

  return dest.requiredFields.map((field, idx) => {
    if (field.isSearchSelectable)
      throw new Error('query select not implemented yet')

    const fieldValue =
      (props.value || []).find((v) => v.fieldID === field.fieldID)?.value || ''

    return (
      <DestinationInputDirect
        key={idx}
        value={fieldValue}
        config={field}
        destType={props.destType}
        disabled={props.disabled || !dest.enabled}
        onChange={(e) => {
          if (!props.onChange) return

          const newValue = e.target.value || ''
          const newValues = (props.value || [])
            .filter((v) => v.fieldID !== field.fieldID)
            .concat({
              fieldID: field.fieldID,
              value: newValue,
            })

          props.onChange(newValues)
        }}
      />
    )
  })
}
