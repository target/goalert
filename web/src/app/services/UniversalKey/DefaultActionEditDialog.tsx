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
  const [q] = useQuery<{
    integrationKey: IntegrationKey
  }>({
    query,
    variables: {
      keyID: props.keyID,
    },
  })

  const [value, setValue] = useState<ActionInput[]>(
    q.data?.integrationKey.config.defaultActions ?? [],
  )
  const [m, commit] = useMutation(mutation)

  const [hasConfirmed, setHasConfirmed] = useState(false)
  const [hasSubmitted, setHasSubmitted] = useState(false)
  const noActionsNoConf = value.length === 0 && !hasConfirmed
  const errs = useErrorConsumer(m.error)
  const [editAction, setEditAction] = useState('')
  const handleChipClick = (action: ActionInput): void => {
    setEditAction(action.dest.type)
  }
  const form = (
    <UniversalKeyActionsForm
      value={value}
      onChange={setValue}
      editActionId={editAction}
      onChipClick={handleChipClick}
      showList
    />
  )

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
        ).then(() => {
          props.onClose()
        })
      }}
      form={form}
      errors={errs.remainingLegacy()}
      notices={getNotice(hasSubmitted, hasConfirmed, setHasConfirmed)}
    />
  )
}
