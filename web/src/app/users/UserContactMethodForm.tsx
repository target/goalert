import { Checkbox, FormControlLabel, Typography } from '@mui/material'
import Grid from '@mui/material/Grid'
import TextField from '@mui/material/TextField'
import React from 'react'
import { DestinationInput } from '../../schema'
import { FormContainer } from '../forms'
import { renderMenuItem } from '../selection/DisableableMenuItem'
import DestinationField from '../selection/DestinationField'
import { useContactMethodTypes } from '../util/RequireConfig'

export type Value = {
  name: string
  dest: DestinationInput
  statusUpdates: boolean
}

export type UserContactMethodFormProps = {
  value: Value

  nameError?: string
  destTypeError?: string
  destFieldErrors?: Readonly<Record<string, string>>

  disabled?: boolean
  edit?: boolean

  disablePortal?: boolean // for testing, disable portal on select menu

  onChange?: (CMValue: Value) => void
}

export default function UserContactMethodForm(
  props: UserContactMethodFormProps,
): React.JSX.Element {
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
            args: {},
          },
        }
      }}
      optionalLabels
    >
      <Grid container spacing={2}>
        <Grid item xs={12} sm={12} md={6}>
          <TextField
            fullWidth
            name='name'
            label='Name'
            disabled={props.disabled}
            error={!!props.nameError}
            helperText={props.nameError}
            value={value.name}
            onChange={(e) =>
              props.onChange &&
              props.onChange({ ...value, name: e.target.value })
            }
          />
        </Grid>
        <Grid item xs={12} sm={12} md={6}>
          <TextField
            fullWidth
            name='dest.type'
            label='Destination Type'
            select
            error={!!props.destTypeError}
            helperText={props.destTypeError}
            SelectProps={{ MenuProps: { disablePortal: props.disablePortal } }}
            value={value.dest.type}
            onChange={(v) =>
              props.onChange &&
              props.onChange({
                ...value,
                dest: {
                  type: v.target.value as string,
                  args: {},
                },
              })
            }
            disabled={props.disabled || edit}
          >
            {destinationTypes.map((t) =>
              renderMenuItem({
                label: t.name,
                value: t.type,
                disabled: !t.enabled,
                disabledMessage: t.enabled ? '' : 'Disabled by administrator.',
              }),
            )}
          </TextField>
        </Grid>
        <Grid item xs={12}>
          <DestinationField
            destType={value.dest.type}
            disabled={props.disabled || edit}
            fieldErrors={props.destFieldErrors}
            value={value.dest.args || {}}
            onChange={(v) =>
              props.onChange &&
              props.onChange({
                ...value,
                dest: {
                  ...value.dest,
                  args: v,
                },
              })
            }
          />
        </Grid>

        {currentType?.userDisclaimer !== '' && (
          <Grid item xs={12}>
            <Typography variant='caption'>
              {currentType?.userDisclaimer}
            </Typography>
          </Grid>
        )}

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
