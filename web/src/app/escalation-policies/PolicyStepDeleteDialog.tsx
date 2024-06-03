import React from 'react'
import { gql, useMutation, useQuery } from 'urql'
import { nonFieldErrors } from '../util/errutil'
import FormDialog from '../dialogs/FormDialog'
import Spinner from '../loading/components/Spinner'
import { GenericError } from '../error-pages'
import { EscalationPolicyStep } from '../../schema'

const query = gql`
  query ($id: ID!) {
    escalationPolicy(id: $id) {
      id
      steps {
        id
      }
    }
  }
`

const mutation = gql`
  mutation ($input: UpdateEscalationPolicyInput!) {
    updateEscalationPolicy(input: $input)
  }
`

export default function PolicyStepDeleteDialog(props: {
  escalationPolicyID: string
  stepID: string
  onClose: () => void
}): JSX.Element {
  const [{ fetching, data, error }] = useQuery({
    query,
    variables: { id: props.escalationPolicyID },
  })

  const [deleteStepMutationStatus, deleteStepMutation] = useMutation(mutation)

  if (fetching && !data) return <Spinner />
  if (error) return <GenericError error={error.message} />

  // get array of step ids without the step to delete
  const sids = data.escalationPolicy.steps.map(
    (s: EscalationPolicyStep) => s.id,
  )
  const toDel = sids.indexOf(props.stepID)
  sids.splice(toDel, 1)

  return (
    <FormDialog
      title='Are you sure?'
      confirm
      subTitle={
        'This will delete step #' +
        (data.escalationPolicy.steps
          .map((s: EscalationPolicyStep) => s.id)
          .indexOf(props.stepID) +
          1) +
        ' on this escalation policy.'
      }
      loading={deleteStepMutationStatus.fetching}
      errors={nonFieldErrors(deleteStepMutationStatus.error)}
      onClose={props.onClose}
      onSubmit={() => {
        deleteStepMutation(
          {
            input: {
              id: data.escalationPolicy.id,
              stepIDs: sids,
            },
          },
          { additionalTypenames: ['EscalationPolicy'] },
        ).then((result) => {
          if (!result.error) props.onClose()
        })
      }}
    />
  )
}
