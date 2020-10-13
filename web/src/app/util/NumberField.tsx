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

export default function NumberField(props: NumberFieldProps) {
  const { float, min = 0, max = 9000, onChange, value, ...rest } = props

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
        let val = e.target.value
        let num = parse(val)

        if (min) num = Math.max(min, num)
        if (max) num = Math.min(max, num)
        val = num.toString()

        setInputValue(val.replace(/[^0-9.]/g, ''))
        if (!isNaN(num)) onChange(val)
      }}
      inputProps={{ min, max, step: 'any' }}
    />
  )
}
