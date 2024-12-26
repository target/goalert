import React, { useState } from 'react'
import { gql, useMutation } from 'urql'
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
): React.JSX.Element {
  const [value, setValue] = useState<Label>({ key: '', value: '' })

  const [{ error }, commit] = useMutation(mutation)

  return (
    <FormDialog
      title='Set Label Value'
      errors={nonFieldErrors(error)}
      onClose={props.onClose}
      onSubmit={() =>
        commit(
          {
            input: {
              ...value,
              target: { type: 'service', id: props.serviceID },
            },
          },
          {
            additionalTypenames: ['Service'],
          },
        ).then((result) => {
          if (!result.error) props.onClose()
        })
      }
      form={
        <ServiceLabelForm
          errors={fieldErrors(error)}
          value={value}
          onChange={(val: Label) => setValue(val)}
        />
      }
    />
  )
}
