import React, { useEffect, useState } from 'react'
import { Add } from '@mui/icons-material'
import { Button, Grid, Typography } from '@mui/material'
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
  editActionId?: string
  onChipClick?: (action: ActionInput) => void

  disablePortal?: boolean
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
    props.editActionId ? getAction(props.value, props.editActionId) : null,
  )
  const [addError, setAddError] = useState<CombinedError | null>(null)
  const valClient = useClient()
  const errs = useErrorConsumer(addError)
  let actions = props.value

  useEffect(() => {
    if (props.editActionId) {
      setCurrentAction(getAction(props.value, props.editActionId))
    }
  }, [props.editActionId])

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
          destTypeError={errs.getErrorByPath(
            'actionInputValidate.input.dest.type',
          )}
          staticParamErrors={errs.getErrorMap(
            'actionInputValidate.input.dest.args',
          )}
          dynamicParamErrors={errs.getErrorMap(
            'actionInputValidate.input.params',
          )}
        />

        {errs.hasErrors() && (
          <Grid item xs={12}>
            {errs.remainingLegacy().map((e) => (
              <Typography key={e.message} color='error'>
                {e.message}
              </Typography>
            ))}
          </Grid>
        )}

        <Grid
          item
          xs={12}
          sx={{
            display: 'flex',
            // justifyContent: 'flex-end',
            alignItems: 'flex-end',
          }}
        >
          <Button
            fullWidth
            startIcon={<Add />}
            variant='contained'
            color='secondary'
            sx={{ height: 'fit-content' }}
            onClick={() => {
              const input = valueToActionInput(currentAction)

              if (props.editActionId !== '') {
                actions = props.value.filter(
                  (v) => v.dest.type !== props.editActionId,
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
            {props.editActionId ? 'Save Action' : 'Add Action'}
          </Button>
        </Grid>
      </Grid>
    </Grid>
  )
}
