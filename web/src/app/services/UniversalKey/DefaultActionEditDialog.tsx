import React, { useState } from 'react'
import { gql, useMutation, useQuery } from 'urql'
import FormDialog from '../../dialogs/FormDialog'
import { nonFieldErrors } from '../../util/errutil'
import { ActionInput, IntegrationKey } from '../../../schema'
import { getNotice } from './utils'
import DefaultActionForm from './DefaultActionForm'

interface DefaultActionEditDialogProps {
  keyID: string
  onClose: () => void
}

const query = gql`
  query UniversalKeyPage($keyID: ID!) {
    integrationKey(id: $keyID) {
      id
      config {
        defaultActions {
          dest {
            type
            values {
              fieldID
              value
            }
          }
          params {
            paramID
            expr
          }
        }
      }
    }
  }
`

const mutation = gql`
  mutation ($input: UpdateKeyConfigInput!) {
    updateKeyConfig(input: $input)
  }
`

export default function DefaultActionEditDialog(
  props: DefaultActionEditDialogProps,
): JSX.Element {
  const [q] = useQuery<{
    integrationKey: IntegrationKey
  }>({
    query,
    variables: {
      keyID: props.keyID,
    },
  })

  // TODO: fetch single rule via query and set it here
  const [value, setValue] = useState<ActionInput[]>(
    q.data?.integrationKey.config.defaultActions ?? [],
  )
  const [editStatus, commit] = useMutation(mutation)

  const [hasConfirmed, setHasConfirmed] = useState(false)
  const [hasSubmitted, setHasSubmitted] = useState(false)
  const noActionsNoConf = value.length === 0 && !hasConfirmed

  return (
    <FormDialog
      title='Edit Default Actions'
      onClose={props.onClose}
      onSubmit={() => {
        if (noActionsNoConf) {
          setHasSubmitted(true)
          return
        }

        return commit(
          {
            input: {
              keyID: props.keyID,
              defaultActions: value,
            },
          },
          { additionalTypenames: ['IntegrationKey', 'Service'] },
        ).then(() => {
          props.onClose()
        })
      }}
      form={<DefaultActionForm value={value} onChange={setValue} />}
      errors={nonFieldErrors(editStatus.error)}
      notices={getNotice(hasSubmitted, hasConfirmed, setHasConfirmed)}
    />
  )
}
