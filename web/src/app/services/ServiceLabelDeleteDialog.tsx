import React from 'react'
import { gql, useMutation } from 'urql'
import { nonFieldErrors } from '../util/errutil'
import FormDialog from '../dialogs/FormDialog'

const mutation = gql`
  mutation ($input: SetLabelInput!) {
    setLabel(input: $input)
  }
`

export default function ServiceLabelDeleteDialog(props: {
  serviceID: string
  labelKey: string
  onClose: () => void
}): React.JSX.Element {
  const { labelKey, onClose, serviceID } = props

  const [deleteLabelStatus, deleteLabel] = useMutation(mutation)

  return (
    <FormDialog
      title='Are you sure?'
      confirm
      subTitle={`This will delete the label: ${labelKey}`}
      loading={deleteLabelStatus.fetching}
      errors={nonFieldErrors(deleteLabelStatus.error)}
      onClose={onClose}
      onSubmit={() => {
        deleteLabel(
          {
            input: {
              key: labelKey,
              value: '',
              target: {
                type: 'service',
                id: serviceID,
              },
            },
          },
          { additionalTypenames: ['Service'] },
        ).then((res) => {
          if (res.error) return
          props.onClose()
        })
      }}
    />
  )
}
