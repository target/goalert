import React from 'react'
import { gql, useMutation } from '@apollo/client'
import p from 'prop-types'
import FormDialog from '../dialogs/FormDialog'
import { nonFieldErrors } from '../util/errutil'

const mutation = gql`
  mutation($id: ID!) {
    deleteAll(input: [{ id: $id, type: contactMethod }])
  }
`
function UserContactMethodDeleteDialog(props) {
  const { contactMethodID, ...rest } = props

  const [deleteCM, deleteCMStatus] = useMutation(mutation, {
    variables: {
      id: contactMethodID,
    },
    onCompleted: props.onClose,
  })

  return (
    <FormDialog
      title='Are you sure?'
      confirm
      loading={deleteCMStatus.loading}
      errors={nonFieldErrors(deleteCMStatus.error)}
      subTitle='This will delete the contact method.'
      caption='This will also delete any notification rules associated with this contact method.'
      onSubmit={() => deleteCM()}
      {...rest}
    />
  )
}

UserContactMethodDeleteDialog.propTypes = {
  contactMethodID: p.string.isRequired,
  onClose: p.func.isRequired, // passed to FormDialog
}

export default UserContactMethodDeleteDialog
