import React, { useState } from 'react'
import { gql, useMutation, useQuery } from 'urql'
import FormDialog from '../../dialogs/FormDialog'
import UniversalKeyRuleForm from './UniversalKeyRuleForm'
import { nonFieldErrors } from '../../util/errutil'
import { IntegrationKey, KeyRule, KeyRuleInput } from '../../../schema'
import { getNotice } from './utils'

interface UniversalKeyRuleEditDialogProps {
  keyID: string
  ruleID: string
  onClose: () => void
}

const query = gql`
  query UniversalKeyPage($keyID: ID!, $ruleID: ID!) {
    integrationKey(id: $keyID) {
      id
      config {
        oneRule(id: $ruleID) {
          id
          name
          description
          conditionExpr
          actions {
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
  }
`

const mutation = gql`
  mutation ($input: UpdateKeyConfigInput!) {
    updateKeyConfig(input: $input)
  }
`

export default function UniversalKeyRuleCreateDialogProps(
  props: UniversalKeyRuleEditDialogProps,
): JSX.Element {
  const [q] = useQuery<{
    integrationKey: IntegrationKey
    ruleKey: KeyRule
  }>({
    query,
    variables: {
      keyID: props.keyID,
      ruleID: props.ruleID,
    },
  })

  // TODO: fetch single rule via query and set it here
  const [value, setValue] = useState<KeyRuleInput>(
    q.data?.integrationKey.config.oneRule ?? {
      id: '',
      name: '',
      description: '',
      conditionExpr: '',
      actions: [],
    },
  )
  const [editStatus, commit] = useMutation(mutation)

  const [hasConfirmed, setHasConfirmed] = useState(false)
  const [hasSubmitted, setHasSubmitted] = useState(false)
  const noActionsNoConf = value.actions.length === 0 && !hasConfirmed

  return (
    <FormDialog
      title='Edit Rule'
      maxWidth='lg'
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
              setRule: {
                id: value.id,
                name: value.name,
                description: value.description,
                conditionExpr: value.conditionExpr,
                actions: value.actions,
              },
            },
          },
          { additionalTypenames: ['IntegrationKey', 'Service'] },
        ).then(() => {
          props.onClose()
        })
      }}
      form={<UniversalKeyRuleForm value={value} onChange={setValue} />}
      errors={nonFieldErrors(editStatus.error)}
      notices={getNotice(hasSubmitted, hasConfirmed, setHasConfirmed)}
    />
  )
}
