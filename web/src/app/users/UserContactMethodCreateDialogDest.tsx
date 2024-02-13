import React, { useState } from 'react'
import { useMutation, useQuery, gql } from 'urql'

import { useErrorsForDest } from '../util/errutil'
import FormDialog from '../dialogs/FormDialog'
import UserContactMethodForm from './UserContactMethodFormDest'
import { useContactMethodTypes } from '../util/RequireConfig'
import { Dialog, DialogTitle, DialogActions, Button } from '@mui/material'
import DialogContentError from '../dialogs/components/DialogContentError'
import { DestinationInput } from '../../schema'

type Value = {
  name: string
  dest: DestinationInput
  statusUpdates: boolean
}

const createMutation = gql`
  mutation CreateUserContactMethodInput($input: CreateUserContactMethodInput!) {
    createUserContactMethod(input: $input) {
      id
    }
  }
`

const userConflictQuery = gql`
  query UserConflictCheck($input: UserSearchOptions) {
    users(input: $input) {
      nodes {
        id
        name
      }
    }
  }
`

const noSuspense = { suspense: false }

export default function UserContactMethodCreateDialogDest(props: {
  userID: string
  onClose: (contactMethodID?: string) => void
  title?: string
  subtitle?: string

  disablePortal?: boolean
}): React.ReactNode {
  const defaultType = useContactMethodTypes()[0] // will be sorted by priority, and enabled first

  // values for contact method form
  const [CMValue, setCMValue] = useState<Value>({
    name: '',
    dest: {
      type: defaultType.type,
      values: [],
    },
    statusUpdates: false,
  })

  const [{ data, fetching: queryLoading }] = useQuery({
    query: userConflictQuery,
    variables: {
      input: {
        dest: CMValue.dest,
      },
    },
    pause:
      !CMValue.dest ||
      !CMValue.dest.values.length ||
      !CMValue.dest.values[0].value,
    context: noSuspense,
  })

  const [createCMStatus, createCM] = useMutation(createMutation)

  const [destTypeErr, destFieldErrs, otherErrs] = useErrorsForDest(
    createCMStatus.error,
    CMValue.dest.type,
    'createUserContactMethod.input.dest',
  )

  if (!defaultType.enabled) {
    // default type will be the first enabled type, so if it's not enabled, none are enabled
    return (
      <Dialog
        disablePortal={props.disablePortal}
        open
        onClose={() => props.onClose()}
      >
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

  const { title = 'Create New Contact Method', subtitle } = props

  const form = (
    <UserContactMethodForm
      disabled={createCMStatus.fetching}
      fieldErrors={destFieldErrs}
      typeError={destTypeErr}
      onChange={(CMValue: Value) => setCMValue(CMValue)}
      value={CMValue}
    />
  )

  return (
    <FormDialog
      disablePortal={props.disablePortal}
      data-cy='create-form'
      title={title}
      subTitle={subtitle}
      loading={createCMStatus.fetching}
      errors={otherErrs}
      onClose={props.onClose}
      // wrapped to prevent event from passing into createCM
      onSubmit={() =>
        createCM(
          {
            input: {
              name: CMValue.name,
              dest: CMValue.dest,
              enableStatusUpdates: CMValue.statusUpdates,
              userID: props.userID,
              newUserNotificationRule: {
                delayMinutes: 0,
              },
            },
          },
          { additionalTypenames: ['UserContactMethod', 'User'] },
        ).then((result) => {
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
