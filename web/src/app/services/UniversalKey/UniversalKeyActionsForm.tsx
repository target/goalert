import React, { useEffect, useState } from 'react'
import { Add, Restore } from '@mui/icons-material'
import { Button, ButtonGroup, Grid, Typography } from '@mui/material'
import { ActionInput } from '../../../schema'
import DynamicActionForm, {
  actionInputToValue,
  valueToActionInput,
} from '../../selection/DynamicActionForm'
import { CombinedError, gql, useClient } from 'urql'
import { useErrorConsumer } from '../../util/ErrorConsumer'
import UniversalKeyActionsList from './UniversalKeyActionsList'

type FormValue = React.ComponentProps<typeof DynamicActionForm>['value']

export type UniversalKeyActionsFormProps = {
  value: Array<ActionInput>
  onChange: (value: Array<ActionInput>) => void

  showList?: boolean
  actionType?: string
  onChipClick?: (action: ActionInput) => void

  disablePortal?: boolean
  setShowNextTooltip: (bool: boolean) => void
}

const query = gql`
  query ValidateActionInput($input: ActionInput!) {
    actionInputValidate(input: $input)
  }
`

const getAction = (actions: ActionInput[], id: string): FormValue => {
  let input
  if ((input = actions.find((v) => v.dest.type === id))) {
    return actionInputToValue(input)
  }
  return null
}

/** Manages a set of actions. */
export default function UniversalKeyActionsForm(
  props: UniversalKeyActionsFormProps,
): React.ReactNode {
  const [currentAction, setCurrentAction] = useState<FormValue>(
    props.actionType ? getAction(props.value, props.actionType) : null,
  )
  const [addError, setAddError] = useState<CombinedError | null>(null)
  const valClient = useClient()
  const errs = useErrorConsumer(addError)
  let actions = props.value

  useEffect(() => {
    if (props.actionType) {
      setCurrentAction(getAction(props.value, props.actionType))
    }
  }, [props.actionType])

  useEffect(() => {
    if (currentAction) {
      console.log('setting tooltip: true')
      props.setShowNextTooltip(true)
    }
  }, [currentAction])

  const destError = errs.getErrorByPath('actionInputValidate.input.dest.type')
  const staticErrors = errs.getErrorMap('actionInputValidate.input.dest.args')
  const dynamicErrors = errs.getErrorMap('actionInputValidate.input.params')

  return (
    <Grid item xs={12} container spacing={2}>
      {props.showList && (
        <UniversalKeyActionsList
          actions={props.value}
          onDelete={(a) => props.onChange(props.value.filter((v) => v !== a))}
          onChipClick={(a) => {
            if (props.onChipClick) {
              props.onChipClick(a)
            }
            setCurrentAction(getAction(props.value, a.dest.type))
          }}
        />
      )}

      <Grid item xs={12} container spacing={2}>
        <DynamicActionForm
          disablePortal={props.disablePortal}
          value={currentAction}
          onChange={setCurrentAction}
          destTypeError={destError}
          staticParamErrors={staticErrors}
          dynamicParamErrors={dynamicErrors}
        />
      </Grid>

      {currentAction?.destType && (
        <Grid
          item
          xs={12}
          sx={{
            display: 'flex',
            alignItems: 'flex-end',
          }}
        >
          <ButtonGroup variant='contained' color='secondary' fullWidth>
            <Button
              startIcon={<Restore />}
              onClick={() => setCurrentAction(null)}
              sx={{ width: '30%' }}
            >
              Reset
            </Button>
            <Button
              type='button'
              startIcon={<Add />}
              onClick={() => {
                const input = valueToActionInput(currentAction)

                if (props.actionType !== '') {
                  actions = props.value.filter(
                    (v) => v.dest.type !== props.actionType,
                  )
                }

                let cancel = ''
                actions.forEach((_a) => {
                  const a = JSON.stringify(_a.dest.args)
                  const cur = JSON.stringify(input.dest.args)
                  if (a === cur) {
                    cancel = 'Cannot add same destination twice'
                  }
                })

                if (cancel !== '') {
                  setAddError({
                    message: cancel,
                  } as CombinedError)
                  return
                }

                // validating input of action
                setAddError(null)
                valClient
                  .query(query, { input })
                  .toPromise()
                  .then((res) => {
                    if (res.error) {
                      setAddError(res.error)
                      return
                    }

                    // clear the current action
                    setCurrentAction(null)
                    props.setShowNextTooltip(false)

                    props.onChange(actions.concat(input))

                    if (props.onChipClick) {
                      props.onChipClick({
                        dest: { type: '', args: {} },
                        params: {},
                      })
                    }
                  })
              }}
            >
              {props.actionType ? 'Save Action' : 'Add Action'}
            </Button>
          </ButtonGroup>
        </Grid>
      )}

      {errs.hasErrors() && (
        <Grid item xs={12}>
          {errs.remainingLegacy().map((err) => (
            <Typography key={err.message} color='error'>
              {err.message}
            </Typography>
          ))}
        </Grid>
      )}
    </Grid>
  )
}
