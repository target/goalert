import React, { useEffect, useState } from 'react'
import { useMutation, gql, CombinedError } from 'urql'

import FormDialog from '../dialogs/FormDialog'
import UserContactMethodForm from './UserContactMethodForm'
import { useContactMethodTypes } from '../util/RequireConfig'
import { Dialog, DialogTitle, DialogActions, Button } from '@mui/material'
import DialogContentError from '../dialogs/components/DialogContentError'
import { DestinationInput } from '../../schema'
import { useErrorConsumer } from '../util/ErrorConsumer'

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

export default function UserContactMethodCreateDialog(props: {
  userID: string
  onClose: (contactMethodID?: string) => void
  title?: string
  subtitle?: string

  disablePortal?: boolean
}): React.ReactNode {
  const defaultType = useContactMethodTypes()[0] // will be sorted by priority, and enabled first

  // values for contact method form
  const [CMValue, _setCMValue] = useState<Value>({
    name: '',
    dest: {
      type: defaultType.type,
      values: [],
    },
    statusUpdates: false,
  })
  const [createErr, setCreateErr] = useState<CombinedError | null>(null)
  const setCMValue = (newValue: Value): void => {
    _setCMValue(newValue)
    setCreateErr(null)
  }

  // TODO: useQuery for userConflictQuery

  const [createCMStatus, createCM] = useMutation(createMutation)
  useEffect(() => {
    setCreateErr(createCMStatus.error || null)
  }, [createCMStatus.error])

  const errs = useErrorConsumer(createErr)

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
      nameError={errs.getError('createUserContactMethod.input.name')}
      destTypeError={errs.getError('createUserContactMethod.input.dest.type')}
      destFieldErrors={errs.getAllDestFieldErrors()}
      onChange={(CMValue: Value) => setCMValue(CMValue)}
      value={CMValue}
      disablePortal={props.disablePortal}
    />
  )

  return (
    <FormDialog
      disablePortal={props.disablePortal}
      data-cy='create-form'
      title={title}
      subTitle={subtitle}
      loading={createCMStatus.fetching}
      errors={errs.remainingLegacy()}
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
