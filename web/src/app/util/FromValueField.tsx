import React, { useEffect, useState } from 'react'
import TextField, { TextFieldProps } from '@material-ui/core/TextField'
import TelTextField from './TelTextField'
import ClickableText from './ClickableText'
import ToggleIcon from '@material-ui/icons/CompareArrows'

export default function FromValueField(
  props: TextFieldProps & { value: string },
): JSX.Element {
  const [phoneMode, setPhoneMode] = useState(!props.value.startsWith('M'))
  useEffect(() => {
    if (props.value.startsWith('M')) {
      setPhoneMode(false)
    } else if (props.value.startsWith('+')) {
      setPhoneMode(true)
    }
  }, [props.value])

  if (!phoneMode) {
    return (
      <TextField
        {...props}
        label='Messaging Service SID'
        InputLabelProps={{
          shrink: true,
        }}
        onChange={(e) => {
          if (!props.onChange) return

          e.target.value = e.target.value.trim().toLowerCase()

          if (e.target.value === 'm') {
            e.target.value = 'M'
          } else if (e.target.value === 'mg') {
            e.target.value = 'MG'
          } else if (e.target.value.startsWith('mg')) {
            e.target.value = 'MG' + e.target.value.replace(/[^0-9a-f]/g, '')
          } else {
            e.target.value = ''
          }

          props.onChange(e)
        }}
        helperText={
          <ClickableText
            data-cy='toggle-duration-off'
            endIcon={<ToggleIcon />}
            onClick={(_e: unknown) => {
              setPhoneMode(true)
              if (!props.onChange) return

              const e = _e as React.ChangeEvent<HTMLInputElement>
              e.target.value = ''
              props.onChange(e)
            }}
          >
            Use a phone number
          </ClickableText>
        }
      />
    )
  }

  return (
    <TelTextField
      {...props}
      label='From Number'
      helperText={
        <ClickableText
          data-cy='toggle-duration-off'
          endIcon={<ToggleIcon />}
          onClick={(_e: unknown) => {
            setPhoneMode(false)
            if (!props.onChange) return

            const e = _e as React.ChangeEvent<HTMLInputElement>
            e.target.value = ''
            props.onChange(e)
          }}
        >
          Use a Messaging Service SID
        </ClickableText>
      }
    />
  )
}
