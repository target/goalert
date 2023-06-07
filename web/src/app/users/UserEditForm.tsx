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
  admin: boolean
  disable: boolean
  passwordRequired: boolean
  username: boolean
  onChange: (newValue: Value) => void
}

function UserEditForm(props: UserEditFormProps): JSX.Element {
  const {
    value,
    errors,
    admin,
    disable,
    passwordRequired,
    username,
    onChange,
  } = props

  const usernameDisabled = disable || !username

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
        {passwordRequired && (
          <Grid item xs={12}>
            <FormField
              fullWidth
              component={TextField}
              name='oldPassword'
              label='Old Password'
              type='password'
              autoComplete={disable ? 'off' : 'current-password'}
              disabled={disable}
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
            autoComplete={disable ? 'off' : 'new-password'}
            disabled={disable}
          />
        </Grid>
        <Grid item xs={12}>
          <FormField
            fullWidth
            component={TextField}
            name='confirmNewPassword'
            label='Confirm New Password'
            type='password'
            autoComplete={disable ? 'off' : 'new-password'}
            disabled={disable}
          />
        </Grid>
        {admin && (
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
