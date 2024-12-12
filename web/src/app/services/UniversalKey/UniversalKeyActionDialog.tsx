import React, { useState } from 'react'
import { gql, useMutation, useQuery } from 'urql'
import { ActionInput, IntegrationKey } from '../../../schema'
import FormDialog from '../../dialogs/FormDialog'
import DynamicActionForm, { Value } from '../../selection/DynamicActionForm'

type UniversalKeyActionDialogProps = {
  keyID: string

  /* The rule ID to add or edit an action for. If not set, operate on the default actions. */
  ruleID?: string

  /* The action index to edit. If not set, add a new action. */
  actionIndex?: number

  onClose: () => void
  disablePortal?: boolean
}

const query = gql`
  query GetKey($keyID: ID!) {
    integrationKey(id: $keyID) {
      id
      config {
        rules {
          id
          name
          actions {
            dest {
              type
              args
            }
            params
          }
        }
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

const updateRule = gql`
  mutation UpdateRule($input: UpdateRuleInput!) {
    updateRule(input: $input) {
      id
    }
  }
`
const updateDefaultActions = gql`
  mutation UpdateDefaultActions($input: UpdateDefaultActionsInput!) {
    updateDefaultActions(input: $input) {
      id
    }
  }
`

export function UniversalKeyActionDialog(
  props: UniversalKeyActionDialogProps,
): React.ReactNode {
  const [q] = useQuery<{ integrationKey: IntegrationKey }>({
    query,
    variables: { keyID: props.keyID },
  })
  if (q.error) throw q.error
  const rule =
    q.data?.integrationKey.config.rules.find((r) => r.id === props.ruleID) ||
    null
  const action = props.actionIndex
    ? rule?.actions[props.actionIndex] ||
      q.data?.integrationKey.config.defaultActions[props.actionIndex]
    : null
  const [value, setValue] = useState<Value>({
    destType: action?.dest.type || '',
    staticParams: action?.dest.args || {},
    dynamicParams: action?.params || {},
  })
  const [m, commit] = useMutation(
    props.ruleID ? updateRule : updateDefaultActions,
  )

  const verb = props.actionIndex ? 'Edit' : 'Add'
  const title = `${verb} ${rule ? '' : 'Default '}Action${rule?.name ? ` for Rule "${rule.name}"` : ''}`

  return (
    <FormDialog
      title={title}
      onClose={props.onClose}
      loading={m.fetching}
      maxWidth='md'
      errors={null}
      onSubmit={() =>
        commit(
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
      }
      form={
        <DynamicActionForm
          disablePortal={props.disablePortal}
          value={value}
          onChange={setValue}
          destTypeError={undefined}
          staticParamErrors={{}}
          dynamicParamErrors={{}}
        />
      }
      PaperProps={{
        sx: {
          minHeight: '500px',
        },
      }}
    />
  )
}
