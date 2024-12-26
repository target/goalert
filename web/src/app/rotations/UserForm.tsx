import React from 'react'
import { FormContainer, FormField } from '../forms'
import { UserSelect } from '../selection'

interface UserFormValue {
  users: Array<string>
}

interface UserFormProps {
  errors?: Array<Error>
  onChange?: (value: UserFormValue) => void
  disabled?: boolean
  value: UserFormValue
}

export default function UserForm(props: UserFormProps): React.JSX.Element {
  return (
    <FormContainer {...props}>
      <FormField
        component={UserSelect}
        disabled={false}
        fieldName='users'
        fullWidth
        label='Select User(s)'
        multiple
        name='users'
        required
        value={props.value.users}
      />
    </FormContainer>
  )
}
