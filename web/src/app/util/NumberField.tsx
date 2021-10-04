import React, { useEffect, useState } from 'react'
import { TextField, TextFieldProps } from '@mui/material'

type NumberFieldProps = TextFieldProps & {
  // float indicates that decimals should be accepted
  float?: boolean
  min?: number
  max?: number
  step?: number | 'any'
  value: string
  onChange: (
    e: React.ChangeEvent<HTMLTextAreaElement | HTMLInputElement>,
  ) => void
}

export default function NumberField(props: NumberFieldProps): JSX.Element {
  const { float, min, max, step = 'any', onChange, value, ...rest } = props

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
      onBlur={(e) => {
        let num = parse(inputValue)
        if (typeof min === 'number' && min > num) num = min
        if (typeof max === 'number' && max < num) num = max
        if (Number.isNaN(num)) {
          // invalid, so revert to actual value
          setInputValue(value.toString())
          return
        }

        // no change
        if (num.toString() === inputValue) return

        // fire change event to clamped value
        setInputValue(num.toString())
        e.target.value = num.toString()
        onChange(e)
      }}
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
      inputProps={{ min, max, step }}
    />
  )
}
