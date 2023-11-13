import React, { useState } from 'react'
import { gql, useMutation } from '@apollo/client'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import FormDialog from '../dialogs/FormDialog'
import { FormContainer, FormField } from '../forms'
import { Checkbox, FormControlLabel, Grid, TextField } from '@mui/material'
import { useConfigValue } from '../util/RequireConfig'
import { useLocation } from 'wouter'

const mutation = gql`
  mutation ($input: CreateUserInput!) {
    createUser(input: $input) {
      id
    }
  }
`
interface UserCreateDialogProps {
  onClose: () => void
}

function UserCreateDialog(props: UserCreateDialogProps): React.ReactNode {
  const [, navigate] = useLocation()
  const [value, setValue] = useState({
    username: '',
    password: '',
    password2: '',
    email: '',
    isAdmin: false,
    name: '',
  })

  const [authDisableBasic] = useConfigValue('Auth.DisableBasic')
  const [createUser, { loading, error }] = useMutation(mutation, {
    variables: {
      input: {
        username: value.username,
        password: value.password,
        name: value.name ? value.name : null,
        email: value.email,
        role: value.isAdmin ? 'admin' : 'user',
        favorite: true,
      },
    },
    onCompleted: (data) => navigate(`/users/${data.createUser.id}`),
  })

  return (
    <FormDialog
      title='Create User'
      loading={loading}
      errors={nonFieldErrors(error)}
      onClose={props.onClose}
      onSubmit={() => createUser()}
      notices={
        authDisableBasic
          ? [
              {
                type: 'WARNING',
                message: 'Basic Auth is Disabled',
                details:
                  'This user will be unable to log in until basic auth is enabled.',
              },
            ]
          : []
      }
      form={
        <FormContainer
          value={value}
          errors={fieldErrors(error)}
          onChange={(val: typeof value) => setValue(val)}
          optionalLabels
        >
          <Grid container spacing={2}>
            <Grid item xs={12}>
              <FormField
                fullWidth
                component={TextField}
                name='username'
                required
              />
            </Grid>
            <Grid item xs={12}>
              <FormField
                fullWidth
                component={TextField}
                name='password'
                type='password'
                autoComplete='new-password'
                required
              />
            </Grid>
            <Grid item xs={12}>
              <FormField
                fullWidth
                component={TextField}
                name='password2'
                label='Confirm Password'
                type='password'
                autoComplete='new-password'
                required
                validate={() => {
                  if (value.password !== value.password2) {
                    return new Error('Passwords do not match')
                  }
                }}
              />
            </Grid>
            <Grid item xs={12}>
              <FormField
                fullWidth
                component={TextField}
                name='name'
                label='Display Name'
              />
            </Grid>
            <Grid item xs={12}>
              <FormField
                fullWidth
                component={TextField}
                name='email'
                type='email'
              />
            </Grid>
            <Grid item xs={12}>
              <FormControlLabel
                control={
                  <FormField component={Checkbox} checkbox name='isAdmin' />
                }
                label='Admin'
                labelPlacement='end'
              />
            </Grid>
          </Grid>
        </FormContainer>
      }
    />
  )
}

export default UserCreateDialog
