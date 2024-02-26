import React, { useEffect, useState } from 'react'
import { useMutation, gql, CombinedError, useQuery } from 'urql'

import { splitErrorsByPath } from '../util/errutil'
import FormDialog from '../dialogs/FormDialog'
import UserContactMethodForm, { errorPaths } from './UserContactMethodFormDest'
import { DestinationInput } from '../../schema'

type Value = {
  name: string
  dest: DestinationInput
  statusUpdates: boolean
}

const query = gql`
  query ($id: ID!) {
    userContactMethod(id: $id) {
      id
      name
      type
      dest {
        type
        values {
          fieldID
          value
          label
        }
      }
      value
      statusUpdates
    }
  }
`

const mutation = gql`
  mutation ($input: UpdateUserContactMethodInput!) {
    updateUserContactMethod(input: $input)
  }
`

export default function UserContactMethodCreateDialogDest(props: {
  onClose: (contactMethodID?: string) => void
  contactMethodID: string
}): React.ReactNode {
  const [{ data, fetching }] = useQuery({
    query,
    variables: { id: props.contactMethodID },
  })

  // values for contact method form
  const [CMValue, _setCMValue] = useState<Value>(data?.userContactMethod)

  const [createErr, setCreateErr] = useState<CombinedError | null>(null)
  const setCMValue = (newValue: Value): void => {
    _setCMValue(newValue)
    setCreateErr(null)
  }

  // TODO: useQuery for userConflictQuery

  const [updateCMStatus, updateCM] = useMutation(mutation)
  useEffect(() => {
    setCreateErr(updateCMStatus.error || null)
  }, [updateCMStatus.error])

  const [formErrors, otherErrs] = splitErrorsByPath(
    createErr,
    errorPaths('createUserContactMethod.input'),
  )

  const form = (
    <UserContactMethodForm
      disabled={updateCMStatus.fetching}
      errors={formErrors}
      onChange={(CMValue: Value) => setCMValue(CMValue)}
      value={CMValue}
    />
  )

  return (
    <FormDialog
      loading={fetching}
      disablePortal={fetching}
      data-cy='edit-form'
      title='Edit Contact Method'
      errors={otherErrs}
      onClose={props.onClose}
      // wrapped to prevent event from passing into createCM
      onSubmit={() => {
        updateCM({
          input: {
            id: props.contactMethodID,
            name: CMValue.name,
            dest: {
              type: CMValue.dest.type,
              values: CMValue.dest.values.map(({ fieldID, value }) => ({
                fieldID,
                value,
              })),
            },
            enableStatusUpdates: Boolean(CMValue.statusUpdates),
          },
        }).then((result) => {
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
