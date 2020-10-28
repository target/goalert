import React, { useEffect, useState } from 'react'
import { TextField, TextFieldProps } from '@material-ui/core'

type NumberFieldProps = TextFieldProps & {
  // float indicates that decimals should be accepted
  float?: boolean
  min?: number
  max?: number
  value: string
  onChange: (val: string) => void
}

export default function NumberField(props: NumberFieldProps): JSX.Element {
  const { float, min, max, onChange, value, ...rest } = props

  const [inputValue, setInputValue] = useState(value)

  useEffect(() => {
    setInputValue(value)
  }, [value])

  const parse = float ? parseFloat : (v: string) => parseInt(v, 10)

  return (
    <TextField
      {...rest}
      value={inputValue}
      type='number'
      onBlur={() => setInputValue(value.toString())}
      onChange={(e) => {
        const val = e.target.value
        const num = parse(val)

        e.target.value = num.toString().replace(/[^0-9.-]/g, '')
        setInputValue(e.target.value)

        if (typeof min === 'number' && min > num) return
        if (typeof max === 'number' && max < num) return
        if (Number.isNaN(num)) return

        return onChange(e)
      }}
      inputProps={{ min, max, step: 'any' }}
    />
  )
}
