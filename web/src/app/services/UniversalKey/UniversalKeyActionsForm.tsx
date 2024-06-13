import React, { useState } from 'react'
import { Add } from '@mui/icons-material'
import { Button, Grid, Typography } from '@mui/material'
import { ActionInput } from '../../../schema'
import DynamicActionForm, {
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
}

const query = gql`
  query ValidateActionInput($input: ActionInput!) {
    actionInputValidate(input: $input)
  }
`

/** Manages a set of actions. */
export default function UniversalKeyActionsForm(
  props: UniversalKeyActionsFormProps,
): React.ReactNode {
  const [currentAction, setCurrentAction] = useState<FormValue>(null)
  const [addError, setAddError] = useState<CombinedError | null>(null)
  const valClient = useClient()
  const errs = useErrorConsumer(addError)

  return (
    <Grid container spacing={2}>
      {props.showList && (
        <UniversalKeyActionsList
          actions={props.value}
          onDelete={(a) => props.onChange(props.value.filter((v) => v !== a))}
        />
      )}

      <Grid item xs={12} container spacing={2}>
        <DynamicActionForm
          value={currentAction}
          onChange={setCurrentAction}
          destTypeError={errs.getError('actionInputValidate.input.dest.type')}
          staticParamErrors={errs.getAllDestFieldErrors(
            'actionInputValidate.input.dest',
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
                  props.onChange(props.value.concat(input))
                })
            }}
          >
            Add Action
          </Button>
        </Grid>
      </Grid>
    </Grid>
  )
}
