import React from 'react'
import { useMutation, gql } from '@apollo/client'
import p from 'prop-types'
import { useHistory } from 'react-router'
import FormDialog from '../dialogs/FormDialog'

const mutation = gql`
  mutation ($input: [TargetInput!]!) {
    deleteAll(input: $input)
  }
`

export default function PolicyDeleteDialog(props) {
  const history = useHistory()
  const [deletePolicy, deletePolicyStatus] = useMutation(mutation, {
    variables: {
      input: [
        {
          type: 'escalationPolicy',
          id: props.escalationPolicyID,
        },
      ],
    },
    onCompleted: () => history.push('/escalation-policies'),
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
