import React from 'react'
import { FormValue } from '../escalation-policies/PolicyStepForm2'
import DestinationInputChip from './DestinationInputChip'
import { DestinationInput } from '../../schema'

export function DestInputChipValueWrapper(): React.ReactNode {
  const [value, setValue] = React.useState<FormValue>({
    delayMinutes: 1,
    actions: [
      {
        type: 'builtin-rotation',
        values: [
          {
            fieldID: 'rotation-id',
            value: 'bf227047-18b8-4de3-881c-24b9dd345670',
          },
        ],
      },
    ],
  })

  function handleDelete(a: DestinationInput): void {
    setValue({
      ...value,
      actions: value.actions.filter((b) => a !== b),
    })
  }

  return value.actions.map((a, idx) => (
    <DestinationInputChip
      key={idx}
      value={a}
      onDelete={() => handleDelete(a)}
    />
  ))
}
