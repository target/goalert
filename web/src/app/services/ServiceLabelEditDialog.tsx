import React, { useState, useEffect } from 'react'
import { gql, useMutation, useQuery } from '@apollo/client'

import { fieldErrors, nonFieldErrors } from '../util/errutil'
import FormDialog from '../dialogs/FormDialog'
import ServiceLabelForm from './ServiceLabelForm'
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

interface ServiceLabelEditDialogProps {
  serviceID: string
  labelKey: string
  onClose: () => void
}

export default function ServiceLabelEditDialog(
  props: ServiceLabelEditDialogProps,
): JSX.Element {
  const { onClose, labelKey, serviceID } = props
  const [value, setValue] = useState({ key: '', value: '' })

  const q = useQuery(query, {
    pollInterval: 0,
    variables: { serviceID },
  })

  const [commit, m] = useMutation(mutation, {
    variables: {
      input: {
        key: value.key,
        value: value.value,
        target: { type: 'service', id: serviceID },
      },
    },
    onCompleted: onClose,
  })

  // load initial form value
  useEffect(() => {
    const initialValue = q.data?.service?.labels?.find(
      (l: Label) => l.key === labelKey,
    )

    if (initialValue && value.key === '') {
      setValue({ key: initialValue.key, value: initialValue.value })
    }
  }, [q.data])

  return (
    <FormDialog
      title='Update Label Value'
      loading={m.loading}
      errors={nonFieldErrors(m.error)}
      onClose={onClose}
      onSubmit={() => commit()}
      form={
        <ServiceLabelForm
          errors={fieldErrors(m.error)}
          editValueOnly
          disabled={q.loading || m.loading}
          value={value}
          onChange={(value) => setValue(value)}
        />
      }
    />
  )
}
