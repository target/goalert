import React from 'react'
import { gql, useQuery, useMutation } from 'urql'
import Spinner from '../loading/components/Spinner'
import FormDialog from '../dialogs/FormDialog'
import { useSessionInfo } from '../util/RequireConfig'
import { GenericError } from '../error-pages'
import { useLocation } from 'wouter'

const query = gql`
  query ($id: ID!) {
    user(id: $id) {
      id
      name
    }
  }
`
const mutation = gql`
  mutation ($input: [TargetInput!]!) {
    deleteAll(input: $input)
  }
`

interface RotationDeleteDialogProps {
  userID: string
  onClose: () => void
}

function UserDeleteDialog(props: RotationDeleteDialogProps): React.ReactNode {
  const { userID: currentUserID, ready: isSessionReady } = useSessionInfo()
  const [, navigate] = useLocation()

  const [{ data, fetching: qLoading, error: qError }] = useQuery({
    query,
    variables: { id: props.userID },
  })

  const [{ fetching: mLoading, error: mError }, deleteUser] =
    useMutation(mutation)

  if (!isSessionReady || (!data && qLoading)) return <Spinner />
  if (qError) return <GenericError error={qError.message} />

  return (
    <FormDialog
      title='Are you sure?'
      confirm
      subTitle={`This will delete the user: ${data?.user?.name}`}
      loading={mLoading}
      errors={mError ? [mError] : []}
      onClose={props.onClose}
      onSubmit={() =>
        deleteUser({
          input: [
            {
              id: props.userID,
              type: 'user',
            },
          ],
        }).then((result) => {
          if (!result.error) return navigate('/users')
        })
      }
      notices={
        props.userID === currentUserID
          ? [
              {
                type: 'WARNING',
                message: 'You are about to delete your own user account',
                details:
                  'You will be logged out immediately and unable to log back in',
              },
            ]
          : []
      }
    />
  )
}

export default UserDeleteDialog
