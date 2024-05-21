import React from 'react'
import { gql, useMutation } from 'urql'
import FormDialog from '../dialogs/FormDialog'
import { nonFieldErrors } from '../util/errutil'

const mutation = gql`
  mutation ($id: ID!) {
    deleteAll(input: [{ id: $id, type: contactMethod }])
  }
`
function UserContactMethodDeleteDialog(props: {
  contactMethodID: string
  onClose: () => void
}): JSX.Element {
  const { contactMethodID, ...rest } = props

  const [deleteCMStatus, deleteCM] = useMutation(mutation)

  return (
    <FormDialog
      title='Are you sure?'
      confirm
      loading={deleteCMStatus.fetching}
      errors={nonFieldErrors(deleteCMStatus.error)}
      subTitle='This will delete the contact method.'
      caption='This will also delete any notification rules associated with this contact method.'
      onSubmit={() =>
        deleteCM(
          {
            id: contactMethodID,
          },
          { additionalTypenames: ['UserContactMethod', 'User'] },
        ).then((res) => {
          if (res.error) return

          props.onClose()
        })
      }
      {...rest}
    />
  )
}

export default UserContactMethodDeleteDialog
