import { Checkbox, FormControlLabel, Typography } from '@mui/material'
import Grid from '@mui/material/Grid'
import TextField from '@mui/material/TextField'
import React from 'react'
import { DestinationInput, StatusUpdateState } from '../../schema'
import { FormContainer, FormField } from '../forms'
import { renderMenuItem } from '../selection/DisableableMenuItem'
import { FieldError } from '../util/errutil'
import DestinationField from '../selection/DestinationField'
import { useContactMethodTypes } from '../util/useDestinationTypes'

type Value = {
  name: string
  dest: DestinationInput
  statusUpdates?: StatusUpdateState
}

export type UserContactMethodFormProps = {
  value: Value

  errors?: Array<FieldError>

  disabled?: boolean
  edit?: boolean

  onChange?: (CMValue: Value) => void
}

export default function UserContactMethodForm(
  props: UserContactMethodFormProps,
): JSX.Element {
  const { value, edit = false, ...other } = props

  const destinationTypes = useContactMethodTypes()
  const currentType = destinationTypes.find((d) => d.type === value.dest.type)

  if (!currentType) throw new Error('invalid destination type')

  const statusUpdateChecked =
    value.statusUpdates === 'ENABLED' ||
    value.statusUpdates === 'ENABLED_FORCED' ||
    false

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
