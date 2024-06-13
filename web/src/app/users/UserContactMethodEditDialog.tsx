import React, { useEffect, useState } from 'react'
import { useMutation, gql, CombinedError, useQuery } from 'urql'

import FormDialog from '../dialogs/FormDialog'
import UserContactMethodForm from './UserContactMethodForm'
import { DestinationInput } from '../../schema'
import { useErrorConsumer } from '../util/ErrorConsumer'

type Value = {
  name: string
  dest: DestinationInput
  statusUpdates: boolean
}

const query = gql`
  query userCm($id: ID!) {
    userContactMethod(id: $id) {
      id
      name
      dest {
        type
        values {
          fieldID
          value
        }
      }
      statusUpdates
    }
  }
`

const mutation = gql`
  mutation UpdateUserContactMethod($input: UpdateUserContactMethodInput!) {
    updateUserContactMethod(input: $input)
  }
`

export default function UserContactMethodEditDialog(props: {
  onClose: (contactMethodID?: string) => void
  contactMethodID: string

  disablePortal?: boolean
}): React.ReactNode {
  const [{ data, fetching }] = useQuery({
    query,
    variables: { id: props.contactMethodID },
  })
  const statusUpdates =
    data?.userContactMethod?.statusUpdates?.includes('ENABLED')
  // values for contact method form
  const [CMValue, _setCMValue] = useState<Value>({
    ...data?.userContactMethod,
    statusUpdates,
  })

  const [updateErr, setUpdateErr] = useState<CombinedError | null>(null)
  const setCMValue = (newValue: Value): void => {
    _setCMValue(newValue)
    setUpdateErr(null)
  }

  const [updateCMStatus, updateCM] = useMutation(mutation)
  useEffect(() => {
    setUpdateErr(updateCMStatus.error || null)
  }, [updateCMStatus.error])
  const errs = useErrorConsumer(updateErr)

  const form = (
    <UserContactMethodForm
      disablePortal={props.disablePortal}
      nameError={errs.getError('createUserContactMethod.input.name')}
      disabled={updateCMStatus.fetching}
      edit
      onChange={(CMValue: Value) => setCMValue(CMValue)}
      value={CMValue}
    />
  )

  return (
    <FormDialog
      loading={fetching}
      disablePortal={props.disablePortal}
      data-cy='edit-form'
      title='Edit Contact Method'
      errors={errs.remainingLegacy()}
      onClose={props.onClose}
      // wrapped to prevent event from passing into createCM
      onSubmit={() => {
        updateCM(
          {
            input: {
              id: props.contactMethodID,
              name: CMValue.name,
              enableStatusUpdates: Boolean(CMValue.statusUpdates),
            },
          },
          { additionalTypenames: ['UserContactMethod'] },
        ).then((result) => {
          if (result.error) {
            return
          }
          props.onClose()
        })
      }}
      form={form}
    />
  )
}
