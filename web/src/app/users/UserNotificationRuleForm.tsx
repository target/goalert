import React from 'react'
import Grid from '@mui/material/Grid'
import TextField from '@mui/material/TextField'
import { FormContainer, FormField } from '../forms'
import UserContactMethodSelect from './UserContactMethodSelect'
import { FieldError } from '../util/errutil'

interface CreateNotificationRule {
  contactMethodID: string
  delayMinutes: number
}

interface UserNotificationRuleFormProps {
  userID: string

  value: CreateNotificationRule

  errors: Error[] | FieldError[]

  onChange: (value: CreateNotificationRule) => void

  disabled: boolean
}

interface Error {
  field: 'delayMinutes' | 'contactMethodID'
  message: string
}

export default function UserNotificationRuleForm(
  props: UserNotificationRuleFormProps,
): React.JSX.Element {
  const { userID, ...other } = props
  return (
    <FormContainer {...other} optionalLabels>
      <Grid container spacing={2}>
        <Grid item xs={12} sm={12} md={6}>
          <FormField
            fullWidth
            name='contactMethodID'
            label='Contact Method'
            userID={userID}
            required
            component={UserContactMethodSelect}
          />
        </Grid>
        <Grid item xs={12} sm={12} md={6}>
          <FormField
            fullWidth
            name='delayMinutes'
            required
            label='Delay (minutes)'
            type='number'
            min={0}
            max={9000}
            component={TextField}
          />
        </Grid>
      </Grid>
    </FormContainer>
  )
}
