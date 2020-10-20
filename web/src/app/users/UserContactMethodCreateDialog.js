import { useMutation, useLazyQuery, gql } from '@apollo/client'
import React, { useState } from 'react'
import p from 'prop-types'

import { fieldErrors, nonFieldErrors } from '../util/errutil'
import FormDialog from '../dialogs/FormDialog'
import UserContactMethodForm from './UserContactMethodForm'

const createMutation = gql`
  mutation($input: CreateUserContactMethodInput!) {
    createUserContactMethod(input: $input) {
      id
    }
  }
`

const userConflictQuery = gql`
  query($input: UserSearchOptions) {
    users(input: $input) {
      nodes {
        id
        name
      }
    }
  }
`

export default function UserContactMethodCreateDialog(props) {
  // values for contact method form
  const [CMValue, setCMValue] = useState({
    name: '',
    type: 'SMS',
    value: '',
  })

  const [query, { data, loading: queryLoading }] = useLazyQuery(
    userConflictQuery,
    {
      variables: {
        input: {
          CMValue: CMValue.value,
          CMType: CMValue.type,
        },
      },
      pollInterval: 0, // override config poll interval to query once
    },
  )

  const [createCM, createCMStatus] = useMutation(createMutation, {
    onCompleted: (result) => {
      props.onClose({ contactMethodID: result.createUserContactMethod.id })
    },
    onError: query,
    variables: {
      input: {
        ...CMValue,
        userID: props.userID,
        newUserNotificationRule: {
          delayMinutes: 0,
        },
      },
    },
  })

  const { loading, error } = createCMStatus
  const { title = 'Create New Contact Method', subtitle } = props

  let fieldErrs = fieldErrors(error)
  if (!queryLoading && data?.users?.nodes?.length > 0) {
    fieldErrs = fieldErrs.map((err) => {
      if (
        err.message === 'contact method already exists for that type and value'
      ) {
        return {
          ...err,
          message: `${err.message}: ${data.users.nodes[0].name}`,
          helpLink: `/users/${data.users.nodes[0].id}`,
        }
      }
      return err
    })
  }

  const form = (
    <UserContactMethodForm
      disabled={loading}
      errors={fieldErrs}
      onChange={(CMValue) => setCMValue(CMValue)}
      value={CMValue}
      disclaimer={props.disclaimer}
    />
  )

  return (
    <FormDialog
      data-cy='create-form'
      title={title}
      subTitle={subtitle}
      loading={loading}
      errors={nonFieldErrors(error)}
      onClose={props.onClose}
      // wrapped to prevent event from passing into createCM
      onSubmit={() => createCM()}
      form={form}
    />
  )
}

UserContactMethodCreateDialog.propTypes = {
  userID: p.string.isRequired,
  onClose: p.func,
  disclaimer: p.string,
  title: p.string,
  subtitle: p.string,
}
