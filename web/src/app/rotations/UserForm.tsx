import React from 'react'
import { FormContainer, FormField } from '../forms'
import { UserSelect } from '../selection'

interface UserFormProps {
  errors?: Array<Error>
  onChange?: (value: { users: string[] }) => void
  disabled?: boolean
  value: { users?: Array<string> }
}

export default function UserForm(props: UserFormProps): JSX.Element {
  return (
    <FormContainer {...props}>
      <FormField
        component={UserSelect}
        disabled={false}
        fieldName='users'
        fullWidth
        label='Select User(s)'
        multiline
        name='users'
        required
        value={props.value.users}
      />
    </FormContainer>
  )
}
