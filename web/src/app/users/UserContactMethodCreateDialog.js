import React, { useState, useEffect } from 'react'
import p from 'prop-types'

import gql from 'graphql-tag'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import Typography from '@material-ui/core/Typography'

import FormDialog from '../dialogs/FormDialog'
import UserContactMethodForm from './UserContactMethodForm'
import { useMutation, useQuery } from '@apollo/react-hooks'

import { AppLink } from '../util/AppLink'

import { useURLParam, useResetURLParams } from '../actions'

const createMutation = gql`
  mutation($input: CreateUserContactMethodInput!) {
    createUserContactMethod(input: $input) {
      id
    }
  }
`

const query = gql`
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

  const [submitted, setSubmitted] = useState(false)

  const { data, loading: queryLoading } = useQuery(query, {
    variables: {
      input: {
        CMValue: CMValue.value,
        CMType: CMValue.type,
      },
    },
  })

  const [userConflict, setUserConflict] = useURLParam('userConflict', '')

  const resetUserConflict = useResetURLParams('userConflict')

  useEffect(() => {
    if (data && !queryLoading) {
      setUserConflict(
        data.users?.nodes[0]?.id + ',' + data.users?.nodes[0]?.name,
      )
    }
  }, [queryLoading])

  const userConflictSplit = userConflict.split(',')
  const existingCMName = userConflictSplit[1]
  const existingCMId = userConflictSplit[0]
  let errorLink = null

  if (existingCMName && existingCMId && submitted) {
    errorLink = (
      <Typography>
        <AppLink to={`/users/${existingCMId}`}>
          Contact method already exists for that type and value:{' '}
          {existingCMName}
        </AppLink>
      </Typography>
    )
  }

  const [createCM, createCMStatus] = useMutation(createMutation, {
    onCompleted: (result) => {
      props.onClose({ contactMethodID: result.createUserContactMethod.id })
    },
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
  const fieldErrs = fieldErrors(error)
  const { title = 'Create New Contact Method', subtitle } = props

  const errors = errorLink ? [errorLink] : nonFieldErrors(error)

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
      errors={errors}
      onClose={() => {
        resetUserConflict()
        props.onClose()
      }}
      // wrapped to prevent event from passing into createCM
      onSubmit={() => {
        // prevent submitting if a user conflict exists
        if (existingCMName && existingCMId) {
          setSubmitted(true)
          return
        }
        createCM()
      }}
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
