import React from 'react'
import { gql, useMutation, useQuery } from '@apollo/client'
import p from 'prop-types'
import { nonFieldErrors } from '../util/errutil'
import FormDialog from '../dialogs/FormDialog'
import Spinner from '../loading/components/Spinner'
import { GenericError } from '../error-pages'

const query = gql`
  query($id: ID!) {
    escalationPolicy(id: $id) {
      id
      steps {
        id
      }
    }
  }
`

const mutation = gql`
  mutation($input: UpdateEscalationPolicyInput!) {
    updateEscalationPolicy(input: $input)
  }
`

function PolicyStepDeleteDialog(props) {
  const { loading, data, error } = useQuery(query, {
    pollInterval: 0,
    variables: { id: props.escalationPolicyID },
  })
  const [deleteStepMutation, deleteStepMutationStatus] = useMutation(mutation, {
    onCompleted: props.onClose,
    update: (cache) => {
      const { escalationPolicy } = cache.readQuery({
        query,
        variables: { id: props.escalationPolicyID },
      })
      cache.writeQuery({
        query,
        variables: { id: data.serviceID },
        data: {
          escalationPolicy: {
            ...escalationPolicy,
            steps: (escalationPolicy.steps || []).filter(
              (step) => step.id !== props.stepID,
            ),
          },
        },
      })
    },
  })

  if (loading && !data) return <Spinner />
  if (error) return <GenericError error={error.message} />

  // get array of step ids without the step to delete
  const sids = data.escalationPolicy.steps.map((s) => s.id)
  const toDel = sids.indexOf(props.stepID)
  sids.splice(toDel, 1)

  return (
    <FormDialog
      title='Are you sure?'
      confirm
      subTitle={
        'This will delete step #' +
        (data.escalationPolicy.steps.map((s) => s.id).indexOf(props.stepID) +
          1) +
        ' on this escalation policy.'
      }
      loading={deleteStepMutationStatus.loading}
      errors={nonFieldErrors(deleteStepMutationStatus.error)}
      onClose={props.onClose}
      onSubmit={() => {
        return deleteStepMutation({
          variables: {
            input: {
              id: data.escalationPolicy.id,
              stepIDs: sids,
            },
          },
        })
      }}
    />
  )
}

PolicyStepDeleteDialog.propTypes = {
  escalationPolicyID: p.string.isRequired,
  stepID: p.string.isRequired,
  onClose: p.func,
}

export default PolicyStepDeleteDialog
