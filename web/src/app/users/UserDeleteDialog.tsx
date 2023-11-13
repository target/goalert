import React from 'react'
import { gql, useQuery, useMutation } from '@apollo/client'
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

  const {
    data,
    loading: qLoading,
    error: qError,
  } = useQuery(query, {
    variables: { id: props.userID },
  })

  const [deleteUser, { loading: mLoading, error: mError }] = useMutation(
    mutation,
    {
      variables: {
        input: [
          {
            id: props.userID,
            type: 'user',
          },
        ],
      },
      onCompleted: ({ deleteAll }) => {
        if (deleteAll) return navigate('/users')
      },
    },
  )

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
      onSubmit={() => deleteUser()}
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
