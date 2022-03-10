import React from 'react'
import { PropTypes as p } from 'prop-types'
import { FormContainer, FormField } from '../forms'
import { UserSelect } from '../selection'

export default function UserForm(props) {
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

UserForm.propTypes = {
  errors: p.array,
  onChange: p.func,
  disabled: p.bool,
  value: p.shape({
    users: p.arrayOf(p.string),
  }).isRequired,
}
