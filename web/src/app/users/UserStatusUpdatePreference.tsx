import React from 'react'
import { useMutation } from '@apollo/client'
import { gql, useQuery } from 'urql'
import UserContactMethodSelect from './UserContactMethodSelect'
import Spinner from '../loading/components/Spinner'
import { GenericError } from '../error-pages'

const query = gql`
  query statusUpdate($id: ID!) {
    user(id: $id) {
      id
      statusUpdateContactMethodID
    }
  }
`
const mutation = gql`
  mutation ($id: ID!, $cmID: ID!) {
    updateUser(input: { id: $id, statusUpdateContactMethodID: $cmID })
  }
`

const disableVal = 'disable'

export default function UserStatusUpdatePreference(props: {
  userID: string
}): JSX.Element {
  const [{ data, error, fetching }] = useQuery({
    query,
    variables: { id: props.userID },
  })
  const [updateCMPreference] = useMutation(mutation)

  if (error) {
    return <GenericError error={error.message} />
  }

  if (fetching && !data) {
    return <Spinner />
  }

  const cmID = data.user.statusUpdateContactMethodID

  return (
    <UserContactMethodSelect
      userID={props.userID}
      label='Alert Status Updates'
      helperText='Update me when my alerts are acknowledged or closed'
      name='alert-status-contact-method'
      value={cmID || disableVal}
      onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
        const contactMethodID =
          e.target.value === disableVal ? '' : e.target.value
        updateCMPreference({
          variables: { id: props.userID, cmID: contactMethodID },
        })
      }}
      extraItems={[{ label: 'Disabled', value: disableVal }]}
    />
  )
}
