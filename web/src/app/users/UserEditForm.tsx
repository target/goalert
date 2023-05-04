import { TextField, FormControlLabel, Checkbox } from '@mui/material'
import Grid from '@mui/material/Grid'
import React from 'react'
import { FormContainer, FormField } from '../forms'
import { FieldError } from '../util/errutil'

export type Value = {
  oldPassword: string
  newPassword: string
  confirmNewPassword: string
  isAdmin: boolean
}

export interface UserEditFormProps {
  value: Value
  errors: Array<FieldError>
  admin: boolean
  onChange: (newValue: Value) => void
}

function UserEditForm(props: UserEditFormProps): JSX.Element {
  const { value, errors, admin, onChange } = props

  return (
    <FormContainer
      value={value}
      errors={errors}
      onChange={(val: Value) => {
        onChange(val)
      }}
    >
      <Grid container spacing={2}>
        {!admin && (
          <Grid item xs={12}>
            <FormField
              InputProps={{
                'data-cy': 'oldPassword',
              }}
              fullWidth
              component={TextField}
              name='oldPassword'
              type='password'
              disabled={admin}
            />
          </Grid>
        )}
        <Grid item xs={12}>
          <FormField
            InputProps={{
              'data-cy': 'newPassword',
            }}
            fullWidth
            component={TextField}
            name='newPassword'
            type='password'
          />
        </Grid>
        <Grid item xs={12}>
          <FormField
            InputProps={{
              'data-cy': 'confirmNewPassword',
            }}
            fullWidth
            component={TextField}
            name='confirmNewPassword'
            label='Confirm New Password'
            type='password'
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
