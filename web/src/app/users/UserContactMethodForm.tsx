import { Checkbox, FormControlLabel, Typography } from '@mui/material'
import Grid from '@mui/material/Grid'
import TextField from '@mui/material/TextField'
import React from 'react'
import { ContactMethodType, StatusUpdateState } from '../../schema'
import { FormContainer, FormField } from '../forms'
import { renderMenuItem } from '../selection/DisableableMenuItem'
import { FieldError } from '../util/errutil'
import DestinationField from '../selection/DestinationField'
import { useContactMethodTypes } from '../util/useDestinationTypes'

type Value = {
  name: string
  type: ContactMethodType
  value: string
  statusUpdates?: StatusUpdateState
}

export type UserContactMethodFormProps = {
  value: Value

  errors?: Array<FieldError>

  disabled?: boolean
  edit?: boolean

  onChange?: (CMValue: Value) => void
}

const isPhoneType = (val: Value): boolean =>
  val.type === 'SMS' || val.type === 'VOICE'

export default function UserContactMethodForm(
  props: UserContactMethodFormProps,
): JSX.Element {
  const { value, edit = false, ...other } = props

  const destinationTypes = useContactMethodTypes()
  const currentType = destinationTypes.find((d) => d.type === value.type)

  const statusUpdateChecked =
    value.statusUpdates === 'ENABLED' ||
    value.statusUpdates === 'ENABLED_FORCED' ||
    false

  return (
    <FormContainer
      {...other}
      value={value}
      mapOnChangeValue={(newValue: Value): Value => {
        // if switching between phone types (or same type), keep the value
        if (
          (isPhoneType(value) && isPhoneType(newValue)) ||
          value.type === newValue.type
        ) {
          return newValue
        }

        return {
          ...newValue,
          value: '',
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
            name='type'
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
            required
            destType={value.type}
            component={DestinationField}
            disabled={edit}
          />
        </Grid>
        <Grid item xs={12}>
          <Typography variant='caption'>
            {currentType?.userDisclaimer}
          </Typography>
        </Grid>
        {edit && (
          <Grid item xs={12}>
            <FormControlLabel
              label='Enable status updates'
              control={
                <Checkbox
                  name='enableStatusUpdates'
                  disabled={
                    value.statusUpdates === 'DISABLED_FORCED' ||
                    value.statusUpdates === 'ENABLED_FORCED'
                  }
                  checked={statusUpdateChecked}
                  onChange={(v) =>
                    props.onChange &&
                    props.onChange({
                      ...value,
                      statusUpdates: v.target.checked ? 'ENABLED' : 'DISABLED',
                    })
                  }
                />
              }
            />
          </Grid>
        )}
      </Grid>
    </FormContainer>
  )
}
