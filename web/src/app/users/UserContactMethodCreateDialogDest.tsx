import React, { useState } from 'react'
import { useMutation, useQuery, gql } from 'urql'

import { fieldErrors, nonFieldErrors } from '../util/errutil'
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
      disablePortal={props.disablePortal}
      data-cy='create-form'
      title={title}
      subTitle={subtitle}
      loading={fetching}
      errors={nonFieldErrors(error)}
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
