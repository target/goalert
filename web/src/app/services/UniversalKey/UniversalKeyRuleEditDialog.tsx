import React, { useState } from 'react'
import { gql, useMutation, useQuery } from 'urql'
import FormDialog from '../../dialogs/FormDialog'
import UniversalKeyRuleForm from './UniversalKeyRuleForm'
import { nonFieldErrors } from '../../util/errutil'
import { IntegrationKey, KeyRule, KeyRuleInput } from '../../../schema'
import { getNotice } from './utils'
import { useErrorConsumer } from '../../util/ErrorConsumer'

interface UniversalKeyRuleEditDialogProps {
  keyID: string
  ruleID: string
  onClose: () => void
  default?: boolean
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
          continueAfterMatch
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
  }>({
    query,
    variables: {
      keyID: props.keyID,
      ruleID: props.ruleID,
    },
  })
  if (q.error) throw q.error

  // shouldn't happen due to suspense
  if (!q.data) throw new Error('failed to load data')

  const rule = q.data.integrationKey.config.oneRule
  if (!rule) throw new Error('rule not found')

  const [value, setValue] = useState<KeyRuleInput>(rule)
  const [m, commit] = useMutation(mutation)

  const [hasConfirmed, setHasConfirmed] = useState(false)
  const [hasSubmitted, setHasSubmitted] = useState(false)
  const noActionsNoConf = value.actions.length === 0 && !hasConfirmed
  const errs = useErrorConsumer(m.error)
  const form = (
    <UniversalKeyRuleForm
      value={value}
      onChange={setValue}
      nameError={errs.getInputError('updateKeyConfig.input.setRule.name')}
      descriptionError={errs.getInputError(
        'updateKeyConfig.input.setRule.description',
      )}
      conditionError={errs.getInputError(
        'updateKeyConfig.input.setRule.conditionExpr',
      )}
    />
  )

  return (
    <FormDialog
      title={props.default ? 'Edit Default Actions' : 'Edit Rule'}
      maxWidth={props.default ? 'sm' : 'lg'}
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
              setRule: value,
            },
          },
          { additionalTypenames: ['KeyConfig'] },
        ).then((res) => {
          if (res.error) return

          props.onClose()
        })
      }}
      form={form}
      errors={errs.remainingLegacy()}
      notices={getNotice(hasSubmitted, hasConfirmed, setHasConfirmed)}
    />
  )
}
