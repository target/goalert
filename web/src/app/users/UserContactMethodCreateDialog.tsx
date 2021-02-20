import React, { useState } from 'react'
import { useMutation, useLazyQuery, gql } from '@apollo/client'

import { fieldErrors, nonFieldErrors } from '../util/errutil'
import FormDialog from '../dialogs/FormDialog'
import UserContactMethodForm from './UserContactMethodForm'
import { useConfigValue } from '../util/RequireConfig'
import { Dialog, DialogTitle, DialogActions, Button } from '@material-ui/core'
import DialogContentError from '../dialogs/components/DialogContentError'
import {
  createNotificationSubscription,
  registerSW,
} from '../util/webpush/webpush'
import { ContactMethodType } from '../../schema'

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

interface UserContactMethodCreateDialogProps {
  userID: string
  onClose: (obj?: { contactMethodID: string }) => void
  disclaimer: string
  title?: string
  subtitle?: string
}

export default function UserContactMethodCreateDialog(
  props: UserContactMethodCreateDialogProps,
): JSX.Element {
  const [allowSV, allowE, allowW, vapidPublicKey] = useConfigValue(
    'Twilio.Enable',
    'SMTP.Enable',
    'WebPushNotifications.Enable',
    'WebPushNotifications.VAPIDPublicKey',
  )
  let typeVal = ''
  if (allowSV) {
    typeVal = 'SMS'
  } else if (allowE) {
    typeVal = 'EMAIL'
  }
  if (allowW) {
    typeVal = 'WEBPUSH'
  }
  // values for contact method form
  const [CMValue, setCMValue] = useState<{
    name: string
    type: ContactMethodType
    value: string
  }>({
    name: '',
    type: typeVal as ContactMethodType,
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

  const [mutateCM, createCMStatus] = useMutation(createMutation, {
    onCompleted: (result) => {
      props.onClose({ contactMethodID: result.createUserContactMethod.id })
    },
    onError: () => query(),
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

  async function createCM(): Promise<void> {
    if (
      CMValue.type === 'WEBPUSH' &&
      window.Notification.permission === 'granted'
    ) {
      // const isRegistered = await isServiceWorkerRegistered('/static/')
      // if (!isRegistered) ...
      const sw = await registerSW()
      const subscription = await createNotificationSubscription(
        sw,
        vapidPublicKey as string,
      )

      const payload = JSON.stringify(subscription)
      console.log(payload)
      setCMValue({ ...CMValue, ...{ value: payload } })
    }

    mutateCM()
  }

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
      onChange={(CMValue: any) => setCMValue(CMValue)}
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
