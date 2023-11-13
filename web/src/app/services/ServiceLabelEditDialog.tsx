import React, { useState } from 'react'
import { gql, useQuery, useMutation } from 'urql'
import { fieldErrors, nonFieldErrors } from '../util/errutil'

import FormDialog from '../dialogs/FormDialog'
import ServiceLabelForm from './ServiceLabelForm'
import Spinner from '../loading/components/Spinner'
import { Label } from '../../schema'

interface Value {
  key: string
  value: string
}

const mutation = gql`
  mutation ($input: SetLabelInput!) {
    setLabel(input: $input)
  }
`
const query = gql`
  query ($serviceID: ID!) {
    service(id: $serviceID) {
      id
      labels {
        key
        value
      }
    }
  }
`

export default function ServiceLabelEditDialog(props: {
  serviceID: string
  labelKey: string
  onClose: () => void
}): React.ReactNode {
  const { onClose, labelKey, serviceID } = props
  const [value, setValue] = useState<Value | null>(null)

  const [{ data, fetching }] = useQuery({
    query,
    variables: { serviceID },
  })

  const [updateLabelStatus, updateLabel] = useMutation(mutation)

  if (!data && fetching) {
    return <Spinner />
  }

  const defaultValue = {
    key: labelKey,
    value: data?.service?.labels?.find((l: Label) => l.key === labelKey).value,
  }

  return (
    <FormDialog
      title='Update Label Value'
      loading={updateLabelStatus.fetching}
      errors={nonFieldErrors(updateLabelStatus.error)}
      onClose={onClose}
      onSubmit={() => {
        if (!value) {
          return onClose()
        }
        updateLabel(
          {
            input: {
              key: labelKey,
              value: value?.value,
              target: { type: 'service', id: serviceID },
            },
          },
          {
            additionalTypenames: ['Service'],
          },
        ).then((res) => {
          if (res.error) return
          props.onClose()
        })
      }}
      form={
        <ServiceLabelForm
          errors={fieldErrors(updateLabelStatus.error)}
          editValueOnly
          disabled={updateLabelStatus.fetching}
          value={value || defaultValue}
          onChange={(value: Value) => setValue(value)}
        />
      }
    />
  )
}
