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

  return dest.requiredFields.map((field) => {
    const fieldValue =
      (props.value || []).find((v) => v.fieldID === field.fieldID)?.value || ''

    function handleChange(newValue: string): void {
      if (!props.onChange) return

      const newValues = (props.value || [])
        .filter((v) => v.fieldID !== field.fieldID)
        .concat({
          fieldID: field.fieldID,
          value: newValue,
        })

      props.onChange(newValues)
    }

    if (field.isSearchSelectable) throw new Error('Not implemented')

    return (
      <DestinationInputDirect
        {...field}
        key={field.fieldID}
        value={fieldValue}
        destType={props.destType}
        disabled={props.disabled || !dest.enabled}
        onChange={(e) => handleChange(e.target.value)}
      />
    )
  })
}
