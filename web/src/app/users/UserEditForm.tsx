import { TextField, FormControlLabel, Checkbox } from '@mui/material'
import Grid from '@mui/material/Grid'
import React from 'react'
import { FormContainer, FormField } from '../forms'
import { FieldError } from '../util/errutil'

export type Value = {
  username: string
  oldPassword: string
  password: string
  confirmNewPassword: string
  isAdmin: boolean
}

export interface UserEditFormProps {
  value: Value
  errors: Array<FieldError>
  isAdmin: boolean
  disabled: boolean
  requireOldPassword: boolean
  hasUsername: boolean
  onChange: (newValue: Value) => void
}

function UserEditForm(props: UserEditFormProps): React.JSX.Element {
  const {
    value,
    errors,
    isAdmin,
    disabled,
    requireOldPassword,
    hasUsername,
    onChange,
  } = props

  const usernameDisabled = disabled || hasUsername

  return (
    <FormContainer
      value={value}
      errors={errors}
      onChange={(val: Value) => {
        onChange(val)
      }}
    >
      <Grid container spacing={2}>
        <Grid item xs={12}>
          <FormField
            fullWidth
            component={TextField}
            autoComplete='off'
            name='username'
            type='username'
            disabled={usernameDisabled}
          />
        </Grid>
        {requireOldPassword && (
          <Grid item xs={12}>
            <FormField
              fullWidth
              component={TextField}
              name='oldPassword'
              label='Old Password'
              type='password'
              autoComplete={disabled ? 'off' : 'current-password'}
              disabled={disabled}
            />
          </Grid>
        )}
        <Grid item xs={12}>
          <FormField
            fullWidth
            component={TextField}
            name='password'
            label='New Password'
            type='password'
            autoComplete={disabled ? 'off' : 'new-password'}
            disabled={disabled}
          />
        </Grid>
        <Grid item xs={12}>
          <FormField
            fullWidth
            component={TextField}
            name='confirmNewPassword'
            label='Confirm New Password'
            type='password'
            autoComplete={disabled ? 'off' : 'new-password'}
            disabled={disabled}
          />
        </Grid>
        {isAdmin && (
          <Grid item xs={12}>
            <FormControlLabel
              control={
                <FormField
                  component={Checkbox}
                  checkbox
                  name='isAdmin'
                  fieldName='isAdmin'
                />
              }
              label='Admin'
              labelPlacement='end'
            />
          </Grid>
        )}
      </Grid>
    </FormContainer>
  )
}

export default UserEditForm
