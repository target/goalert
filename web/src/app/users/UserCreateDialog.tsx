import React, { useState } from 'react'
import { gql, useMutation } from '@apollo/client'
import { Redirect } from 'react-router'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import FormDialog from '../dialogs/FormDialog'
import { FormContainer, FormField } from '../forms'
import { Checkbox, FormControlLabel, Grid, TextField } from '@material-ui/core'
import { useConfigValue } from '../util/RequireConfig'

const mutation = gql`
  mutation ($input: CreateUserInput!) {
    createUser(input: $input) {
      id
    }
  }
`
const initialValue = {
  username: '',
  password: '',
  email: '',
  isAdmin: false,
}

interface UserCreateDialogProps {
  onClose: () => void
}

function UserCreateDialog(props: UserCreateDialogProps): JSX.Element {
  const [value, setValue] = useState(initialValue)
  const [authDisableBasic] = useConfigValue('Auth.DisableBasic')
  const [createUser, { loading, data, error }] = useMutation(mutation, {
    variables: {
      input: {
        username: value.username,
        password: value.password,
        email: value.email,
        role: value.isAdmin ? 'admin' : 'user',
      },
    },
  })

  if (data?.createUser) {
    return <Redirect push to={`/users/${data.createUser.id}`} />
  }

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
        >
          <Grid container spacing={2}>
            <Grid item xs={12}>
              <FormField
                fullWidth
                component={TextField}
                label='Username'
                name='username'
                required
              />
            </Grid>
            <Grid item xs={12}>
              <FormField
                fullWidth
                component={TextField}
                label='Password'
                name='password'
                required
              />
            </Grid>
            <Grid item xs={12}>
              <FormField
                fullWidth
                component={TextField}
                label='Email'
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
