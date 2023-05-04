import React, { useState } from 'react'
import { gql, useMutation } from '@apollo/client'

import { fieldErrors, nonFieldErrors } from '../util/errutil'
import FormDialog from '../dialogs/FormDialog'
import ServiceLabelForm from './ServiceLabelForm'
import { Label } from '../../schema'

const mutation = gql`
  mutation ($input: SetLabelInput!) {
    setLabel(input: $input)
  }
`

interface ServiceLabelCreateDialogProps {
  serviceID: string
  onClose: () => void
}

export default function ServiceLabelCreateDialog(
  props: ServiceLabelCreateDialogProps,
): JSX.Element {
  const [value, setValue] = useState<Label>({ key: '', value: '' })

  const [createLabel, { loading, error }] = useMutation(mutation, {
    variables: {
      input: {
        ...value,
        target: { type: 'service', id: props.serviceID },
      },
    },
    onCompleted: props.onClose,
  })

  return (
    <FormDialog
      title='Set Label Value'
      loading={loading}
      errors={nonFieldErrors(error)}
      onClose={props.onClose}
      onSubmit={() => createLabel()}
      form={
        <ServiceLabelForm
          errors={fieldErrors(error)}
          disabled={loading}
          value={value}
          onChange={(val: Label) => setValue(val)}
        />
      }
    />
  )
}
