import React from 'react'
import { useMutation, gql } from '@apollo/client'
import p from 'prop-types'
import FormDialog from '../dialogs/FormDialog'
import { useLocation } from 'wouter'

const mutation = gql`
  mutation ($input: [TargetInput!]!) {
    deleteAll(input: $input)
  }
`

export default function PolicyDeleteDialog(props) {
  const [, navigate] = useLocation()
  const [deletePolicy, deletePolicyStatus] = useMutation(mutation, {
    variables: {
      input: [
        {
          type: 'escalationPolicy',
          id: props.escalationPolicyID,
        },
      ],
    },
    onCompleted: () => navigate('/escalation-policies'),
  })

  return (
    <FormDialog
      title='Are you sure?'
      confirm
      subTitle='You will not be able to delete this policy if it is in use by one or more services.'
      loading={deletePolicyStatus.loading}
      errors={deletePolicyStatus.error ? [deletePolicyStatus.error] : []}
      onClose={props.onClose}
      onSubmit={() => deletePolicy()}
    />
  )
}

PolicyDeleteDialog.propTypes = {
  escalationPolicyID: p.string.isRequired,
  onClose: p.func,
}
