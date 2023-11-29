import React, { useState } from 'react'
import { useMutation, useQuery, gql } from 'urql'

import { fieldErrors, nonFieldErrors } from '../util/errutil'
import FormDialog from '../dialogs/FormDialog'
import UserContactMethodForm from './UserContactMethodForm'
import { useConfigValue } from '../util/RequireConfig'
import { Dialog, DialogTitle, DialogActions, Button } from '@mui/material'
import DialogContentError from '../dialogs/components/DialogContentError'
import { ContactMethodType } from '../../schema'

type Value = {
  name: string
  type: ContactMethodType
  value: string
}

const createMutation = gql`
  mutation ($input: CreateUserContactMethodInput!) {
    createUserContactMethod(input: $input) {
      id
    }
  }
`

const userConflictQuery = gql`
  query ($input: UserSearchOptions) {
    users(input: $input) {
      nodes {
        id
        name
      }
    }
  }
`

const noSuspense = { suspense: false }

export default function UserContactMethodCreateDialog(props: {
  userID: string
  onClose: (contactMethodID?: string) => void
  title?: string
  subtitle?: string
}): React.ReactNode {
  const [allowSV, allowE, allowW, allowS] = useConfigValue(
    'Twilio.Enable',
    'SMTP.Enable',
    'Webhook.Enable',
    'Slack.Enable',
  )

  let typeVal: ContactMethodType = 'VOICE'
  if (allowSV) {
    typeVal = 'SMS'
  } else if (allowE) {
    typeVal = 'EMAIL'
  } else if (allowW) {
    typeVal = 'WEBHOOK'
  } else if (allowS) {
    typeVal = 'SLACK_DM'
  }

  // values for contact method form
  const [CMValue, setCMValue] = useState<Value>({
    name: '',
    type: typeVal,
    value: '',
  })

  const [{ data, fetching: queryLoading }] = useQuery({
    query: userConflictQuery,
    variables: {
      input: {
        CMValue: CMValue.value,
        CMType: CMValue.type,
      },
    },
    context: noSuspense,
  })

  const [createCMStatus, createCM] = useMutation(createMutation)

  if (!typeVal) {
    return (
      <Dialog open onClose={() => props.onClose()}>
        <DialogTitle>No Contact Types Available</DialogTitle>
        <DialogContentError error='There are no contact types currently enabled by the administrator.' />
        <DialogActions>
          <Button variant='contained' onClick={() => props.onClose()}>
            Okay
          </Button>
        </DialogActions>
      </Dialog>
    )
  }

  const { fetching, error } = createCMStatus
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
      disabled={fetching}
      errors={fieldErrs}
      onChange={(CMValue: Value) => setCMValue(CMValue)}
      value={CMValue}
    />
  )

  return (
    <FormDialog
      data-cy='create-form'
      title={title}
      subTitle={subtitle}
      loading={fetching}
      errors={nonFieldErrors(error)}
      onClose={props.onClose}
      // wrapped to prevent event from passing into createCM
      onSubmit={() =>
        createCM({
          input: {
            ...CMValue,
            userID: props.userID,
            newUserNotificationRule: {
              delayMinutes: 0,
            },
          },
        }).then((result) => {
          if (result.error) {
            return
          }
          props.onClose(result.data.createUserContactMethod.id)
        })
      }
      form={form}
    />
  )
}
