import React from 'react'
import p from 'prop-types'
import gql from 'graphql-tag'
import { useMutation } from '@apollo/react-hooks'
import { push } from 'connected-react-router'
import { useDispatch } from 'react-redux'
import FormDialog from '../dialogs/FormDialog'

const mutation = gql`
  mutation($input: [TargetInput!]!) {
    deleteAll(input: $input)
  }
`

export default function PolicyDeleteDialog(props) {
  const dispatch = useDispatch()
  const [deletePolicy, deletePolicyStatus] = useMutation(mutation, {
    refetchQueries: ['epsQuery'],
    variables: {
      input: [
        {
          type: 'escalationPolicy',
          id: props.escalationPolicyID,
        },
      ],
    },
    onCompleted: () => dispatch(push('/escalation-policies')),
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
