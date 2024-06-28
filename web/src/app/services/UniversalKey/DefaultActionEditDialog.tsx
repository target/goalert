import React, { useState } from 'react'
import { gql, useMutation, useQuery } from 'urql'
import FormDialog from '../../dialogs/FormDialog'
import { ActionInput, IntegrationKey } from '../../../schema'
import { getNotice } from './utils'
import UniversalKeyActionsForm from './UniversalKeyActionsForm'
import { useErrorConsumer } from '../../util/ErrorConsumer'

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
            args
          }
          params
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
  const [keyResult] = useQuery<{
    integrationKey: IntegrationKey
  }>({
    query,
    variables: {
      keyID: props.keyID,
    },
  })

  const [value, setValue] = useState<ActionInput[]>(
    keyResult.data?.integrationKey.config.defaultActions ?? [],
  )
  const [updateKeyResult, commit] = useMutation(mutation)

  const [hasConfirmed, setHasConfirmed] = useState(false)
  const [hasSubmitted, setHasSubmitted] = useState(false)
  const noActionsNoConf = value.length === 0 && !hasConfirmed
  const errs = useErrorConsumer(updateKeyResult.error)
  const [editAction, setEditAction] = useState('')

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
          { additionalTypenames: ['KeyConfig'] },
        ).then((res) => {
          if (!res.error) props.onClose()
        })
      }}
      form={
        <UniversalKeyActionsForm
          value={value}
          onChange={setValue}
          editActionId={editAction}
          onChipClick={(action: ActionInput) => setEditAction(action.dest.type)}
          showList
        />
      }
      errors={errs.remainingLegacy()}
      notices={getNotice(hasSubmitted, hasConfirmed, setHasConfirmed)}
    />
  )
}
