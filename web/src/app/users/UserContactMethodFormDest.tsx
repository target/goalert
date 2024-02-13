import { Checkbox, FormControlLabel, Typography } from '@mui/material'
import Grid from '@mui/material/Grid'
import TextField from '@mui/material/TextField'
import React from 'react'
import { DestinationInput } from '../../schema'
import { FormContainer, FormField } from '../forms'
import { renderMenuItem } from '../selection/DisableableMenuItem'
import { InputFieldError, SimpleError } from '../util/errutil'
import DestinationField from '../selection/DestinationField'
import { useContactMethodTypes } from '../util/RequireConfig'

export type Value = {
  name: string
  dest: DestinationInput
  statusUpdates: boolean
}

export type UserContactMethodFormProps = {
  value: Value

  typeError?: SimpleError
  fieldErrors?: Array<InputFieldError>

  disabled?: boolean
  edit?: boolean

  onChange?: (CMValue: Value) => void
}

export default function UserContactMethodFormDest(
  props: UserContactMethodFormProps,
): JSX.Element {
  const { value, edit = false, ...other } = props

  const destinationTypes = useContactMethodTypes()
  const currentType = destinationTypes.find((d) => d.type === value.dest.type)

  if (!currentType) throw new Error('invalid destination type')

  let statusLabel = 'Send alert status updates'
  let statusUpdateChecked = value.statusUpdates
  if (currentType.statusUpdatesRequired) {
    statusLabel = 'Send alert status updates (cannot be disabled for this type)'
    statusUpdateChecked = true
  } else if (!currentType.supportsStatusUpdates) {
    statusLabel = 'Send alert status updates (not supported for this type)'
    statusUpdateChecked = false
  }

  return (
    <FormContainer
      {...other}
      value={value}
      mapOnChangeValue={(newValue: Value): Value => {
        if (newValue.dest.type === value.dest.type) {
          return newValue
        }

        // reset otherwise
        return {
          ...newValue,
          dest: {
            ...newValue.dest,
            values: [],
          },
        }
      }}
      optionalLabels
    >
      <Grid container spacing={2}>
        <Grid item xs={12} sm={12} md={6}>
          <FormField fullWidth name='name' required component={TextField} />
        </Grid>
        <Grid item xs={12} sm={12} md={6}>
          <FormField
            fullWidth
            name='dest.type'
            required
            select
            disabled={edit}
            component={TextField}
          >
            {destinationTypes.map((t) =>
              renderMenuItem({
                label: t.name,
                value: t.type,
                disabled: !t.enabled,
                disabledMessage: t.enabled ? '' : t.disabledMessage,
              }),
            )}
          </FormField>
        </Grid>
        <Grid item xs={12}>
          <FormField
            fullWidth
            name='value'
            fieldName='dest.values'
            required
            destType={value.dest.type}
            component={DestinationField}
            disabled={edit}
            destFieldErrors={props.fieldErrors}
          />
        </Grid>
        <Grid item xs={12}>
          <Typography variant='caption'>
            {currentType?.userDisclaimer}
          </Typography>
        </Grid>

        <Grid item xs={12}>
          <FormControlLabel
            label={statusLabel}
            title='Alert status updates are sent when an alert is acknowledged, closed, or escalated.'
            control={
              <Checkbox
                name='enableStatusUpdates'
                disabled={
                  !currentType.supportsStatusUpdates ||
                  currentType.statusUpdatesRequired ||
                  props.disabled
                }
                checked={statusUpdateChecked}
                onChange={(v) =>
                  props.onChange &&
                  props.onChange({
                    ...value,
                    statusUpdates: v.target.checked,
                  })
                }
              />
            }
          />
        </Grid>
      </Grid>
    </FormContainer>
  )
}
