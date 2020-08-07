import React, { useState } from 'react'
import p from 'prop-types'

import gql from 'graphql-tag'
import { fieldErrors, nonFieldErrors } from '../util/errutil'

import FormDialog from '../dialogs/FormDialog'
import UserContactMethodForm from './UserContactMethodForm'
import { useMutation } from '@apollo/react-hooks'
import { useConfigValue } from '../util/RequireConfig'
import { Dialog, DialogTitle, DialogActions, Button } from '@material-ui/core'
import DialogContentError from '../dialogs/components/DialogContentError'

const createMutation = gql`
  mutation($input: CreateUserContactMethodInput!) {
    createUserContactMethod(input: $input) {
      id
    }
  }
`

export default function UserContactMethodCreateDialog(props) {
  const [allowSV, allowE] = useConfigValue('Twilio.Enable', 'SMTP.Enable')
  let typeVal = ''
  if (allowSV) {
    typeVal = 'SMS'
  } else if (allowE) {
    typeVal = 'EMAIL'
  }
  // values for contact method form
  const [cmValue, setCmValue] = useState({
    name: '',
    type: typeVal,
    value: '',
  })

  const [createCM, createCMStatus] = useMutation(createMutation, {
    onCompleted: (result) => {
      props.onClose({ contactMethodID: result.createUserContactMethod.id })
    },
    variables: {
      input: {
        ...cmValue,
        userID: props.userID,
        newUserNotificationRule: {
          delayMinutes: 0,
        },
      },
    },
  })

  if (!typeVal) {
    return (
      <Dialog open onClose={() => props.onClose()}>
        <DialogTitle>No Contact Types Available</DialogTitle>
        <DialogContentError error='There are no contact types currently enabled by the administrator.' />
        <DialogActions>
          <Button
            color='primary'
            variant='contained'
            onClick={() => props.onClose()}
          >
            Okay
          </Button>
        </DialogActions>
      </Dialog>
    )
  }

  const { loading, error } = createCMStatus
  const fieldErrs = fieldErrors(error)
  const { title = 'Create New Contact Method', subtitle } = props

  const form = (
    <UserContactMethodForm
      disabled={loading}
      errors={fieldErrs}
      onChange={(cmValue) => setCmValue(cmValue)}
      value={cmValue}
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
