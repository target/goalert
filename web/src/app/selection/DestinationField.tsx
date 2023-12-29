import React from 'react'
import { DestinationType, FieldValueInput } from '../../schema'
import DestinationInputDirect from './DestinationInputDirect'
import { useDestinationType } from '../util/useDestinationTypes'
import DestinationSearchSelect from './DestinationSearchSelect'

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

    if (field.isSearchSelectable)
      return (
        <DestinationSearchSelect
          key={field.fieldID}
          value={fieldValue}
          config={field}
          destType={props.destType}
          disabled={props.disabled || !dest.enabled}
          onChange={handleChange}
        />
      )

    return (
      <DestinationInputDirect
        key={field.fieldID}
        value={fieldValue}
        config={field}
        destType={props.destType}
        disabled={props.disabled || !dest.enabled}
        onChange={(e) => handleChange(e.target.value)}
      />
    )
  })
}