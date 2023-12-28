import React from 'react'
import TelTextField from './TelTextField'

export function TestWrapper(): React.ReactNode {
  const [value, setValue] = React.useState('')

  return (
    <TelTextField
      value={value}
      onChange={(v) => {
        setValue(v.target.value)
      }}
    />
  )
}
