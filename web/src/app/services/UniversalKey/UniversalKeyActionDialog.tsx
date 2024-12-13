import React, { useState } from 'react'
import { gql, useMutation, useQuery } from 'urql'
import {
  Action,
  ActionInput,
  IntegrationKey,
  UpdateKeyConfigInput,
} from '../../../schema'
import FormDialog from '../../dialogs/FormDialog'
import DynamicActionForm, { Value } from '../../selection/DynamicActionForm'
import { useDefaultAction } from '../../util/RequireConfig'
import { useErrorConsumer } from '../../util/ErrorConsumer'
import { Checkbox, FormControlLabel, Typography } from '@mui/material'

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

const updateKeyConfig = gql`
  mutation UpdateKeyConfig($input: UpdateKeyConfigInput!) {
    updateKeyConfig(input: $input)
  }
`

function actionToInput(action: Action): ActionInput {
  return {
    dest: action.dest,
    params: action.params,
  }
}

export function UniversalKeyActionDialog(
  props: UniversalKeyActionDialogProps,
): React.ReactNode {
  const defaultAction = useDefaultAction()
  const [q] = useQuery<{ integrationKey: IntegrationKey }>({
    query,
    variables: { keyID: props.keyID },
  })
  if (q.error) throw q.error

  const config = q.data?.integrationKey.config
  if (!config) throw new Error('missing config')

  const rule = config.rules.find((r) => r.id === props.ruleID) || null
  if (props.ruleID && !rule) throw new Error('missing rule')
  const actions = rule ? rule.actions : config.defaultActions
  const action =
    props.actionIndex !== undefined ? actions[props.actionIndex] : null
  const [value, setValue] = useState<Value>({
    destType: action?.dest.type || defaultAction.dest.type,
    staticParams: action?.dest.args || {},
    dynamicParams: action?.params || defaultAction.params,
  })
  const [deleteAction, setDeleteAction] = useState(false)
  const [m, commit] = useMutation(updateKeyConfig)

  const verb = action ? 'Edit' : 'Add'
  const title = `${verb} ${rule ? '' : 'Default '}Action${rule?.name ? ` for Rule "${rule.name}"` : ''}`

  const input = { keyID: props.keyID } as UpdateKeyConfigInput
  const newAction = {
    dest: {
      type: value.destType,
      args: value.staticParams,
    },
    params: value.dynamicParams,
  }
  if (rule && props.actionIndex !== undefined) {
    // Edit rule action
    input.setRuleActions = {
      id: rule.id,
      actions: actions
        .map(actionToInput)
        .map((a, idx) => (idx === props.actionIndex ? newAction : a))
        .filter((a, idx) => idx !== props.actionIndex || !deleteAction), // remove action if deleteAction
    }
  } else if (rule) {
    // Add rule action
    input.setRuleActions = {
      id: rule.id,
      actions: actions.map(actionToInput).concat(newAction),
    }
  } else if (props.actionIndex !== undefined) {
    // Edit default action
    input.defaultActions = actions
      .map(actionToInput)
      .map((a, idx) => (idx === props.actionIndex ? newAction : a))
      .filter((a, idx) => idx !== props.actionIndex || !deleteAction) // remove action if deleteAction
  } else {
    // Add default action
    input.defaultActions = actions.map(actionToInput).concat(newAction)
  }

  const errs = useErrorConsumer(m.error)

  return (
    <FormDialog
      title={title}
      onClose={props.onClose}
      loading={m.fetching}
      maxWidth='md'
      errors={errs.remainingLegacyCallback()}
      onSubmit={() =>
        commit({ input }, { additionalTypenames: ['KeyConfig'] }).then(
          (res) => {
            if (res.error) return

            props.onClose()
          },
        )
      }
      form={
        <React.Fragment>
          {props.actionIndex !== undefined && (
            <FormControlLabel
              sx={{ paddingBottom: '1em' }}
              control={
                <Checkbox
                  checked={deleteAction}
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                    setDeleteAction(e.target.checked)
                  }
                />
              }
              label='Delete this action'
            />
          )}
          {!deleteAction && (
            <DynamicActionForm
              disablePortal={props.disablePortal}
              value={value}
              onChange={setValue}
              staticParamErrors={errs.getErrorMap('updateKeyConfig')}
              dynamicParamErrors={errs.getErrorMap(
                /updateKeyConfig.input.defaultActions.\d+.params/,
              )}
            />
          )}
          {deleteAction && (
            <Typography>
              This will remove the action from{' '}
              {rule ? 'the rule' : 'default actions'}.
            </Typography>
          )}
        </React.Fragment>
      }
    />
  )
}
