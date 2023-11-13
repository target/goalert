import React from 'react'
import { useMutation, gql } from 'urql'
import FormDialog from '../dialogs/FormDialog'
import { useLocation } from 'wouter'

const mutation = gql`
  mutation ($input: [TargetInput!]!) {
    deleteAll(input: $input)
  }
`

export default function PolicyDeleteDialog(props: {
  escalationPolicyID: string
  onClose: () => void
}): React.ReactNode {
  const [, navigate] = useLocation()
  const [deletePolicyStatus, deletePolicy] = useMutation(mutation)

  return (
    <FormDialog
      title='Are you sure?'
      confirm
      subTitle='You will not be able to delete this policy if it is in use by one or more services.'
      loading={deletePolicyStatus.fetching}
      errors={deletePolicyStatus.error ? [deletePolicyStatus.error] : []}
      onClose={props.onClose}
      onSubmit={() =>
        deletePolicy(
          {
            input: [
              {
                type: 'escalationPolicy',
                id: props.escalationPolicyID,
              },
            ],
          },
          { additionalTypenames: ['EscalationPolicy'] },
        ).then((result) => {
          if (!result.error) {
            navigate('/escalation-policies')
          }
        })
      }
    />
  )
}
