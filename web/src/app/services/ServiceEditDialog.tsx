import React, { useState } from 'react'
import { gql, useQuery, useMutation } from 'urql'
import { fieldErrors, nonFieldErrors } from '../util/errutil'

import FormDialog from '../dialogs/FormDialog'
import ServiceLabelForm from './ServiceLabelForm'
import Spinner from '../loading/components/Spinner'
import { Label } from '../../schema'

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
}): JSX.Element {
  const { onClose, labelKey, serviceID } = props
  const [value, setValue] = useState<Label | null>(null)
  const [{ data, fetching }] = useQuery({
    query,
    variables: {
      serviceID: props.serviceID,
    },
  })

  const [updateLabelStatus, updateLabel] = useMutation(mutation)

  if (fetching || !data) return <Spinner />

  const defaultValue = {
    key: data.service.labels.find((l: Label) => l.key === labelKey).key,
    value: data.service.labels.find((l: Label) => l.key === labelKey).value,
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
              ...value,
              target: { type: 'service', id: serviceID },
            },
          },
          { additionalTypenames: ['Service'] },
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
          value={value || { key: defaultValue.key, value: defaultValue.value }}
          onChange={(value) => setValue(value)}
        />
      }
    />
  )
}
