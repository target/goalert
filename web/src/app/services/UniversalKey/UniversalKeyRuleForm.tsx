import React, { useState } from 'react'
import { FormContainer, FormField } from '../../forms'
import { Button, Grid, TextField, Typography } from '@mui/material'
import { ActionInput, KeyRuleInput } from '../../../schema'
import { useDynamicActionTypes } from '../../util/RequireConfig'
import DestinationInputChip from '../../util/DestinationInputChip'
import { gql, useClient } from 'urql'
import DynamicActionField, {
  Value as ActionValue,
  defaults,
  valueToActionInput,
} from '../../selection/DynamicActionField'
import { Add } from '@mui/icons-material'

interface UniversalKeyRuleFormProps {
  value: KeyRuleInput
  onChange: (val: KeyRuleInput) => void
}

const query = gql`
  query DestDisplayInfo($input: DestinationInput!) {
    destinationDisplayInfo(input: $input) {
      text
      iconURL
      iconAltText
      linkURL
    }
  }
`

export default function UniversalKeyRuleForm(
  props: UniversalKeyRuleFormProps,
): JSX.Element {
  const types = useDynamicActionTypes()

  const [currentAction, setCurrentAction] = useState<ActionValue>(
    defaults(types[0]),
  )

  const validationClient = useClient()

  function handleDelete(a: ActionInput): void {
    if (!props.onChange) return
    props.onChange({
      ...props.value,
      actions: props.value.actions.filter((b) => a !== b),
    })
  }

  return (
    <FormContainer value={props.value} onChange={props.onChange}>
      <Grid container spacing={2}>
        <Grid item xs={12}>
          <FormField
            fullWidth
            component={TextField}
            label='Name'
            name='name'
            required
          />
        </Grid>
        <Grid item xs={12}>
          <FormField
            fullWidth
            component={TextField}
            label='Description'
            name='description'
          />
        </Grid>
        <Grid item xs={12}>
          <FormField
            fullWidth
            component={TextField}
            label='Condition (Expr syntax)'
            name='conditionExpr'
            required
            multiline
            rows={3}
          />
        </Grid>
        <Grid item xs={12} sx={{ mb: -1 }}>
          <Typography variant='h6' color='textPrimary'>
            Actions
          </Typography>
        </Grid>
        <Grid item xs={12} container spacing={1} sx={{ p: 1 }}>
          {props.value.actions.map((a) => (
            <Grid item key={JSON.stringify(a.dest)}>
              <DestinationInputChip
                value={a.dest}
                onDelete={() => handleDelete(a)}
              />
            </Grid>
          ))}
          {props.value.actions.length === 0 && (
            <Grid item xs={12}>
              <Typography variant='body2' color='textSecondary'>
                No actions
              </Typography>
            </Grid>
          )}
        </Grid>

        <DynamicActionField value={currentAction} onChange={setCurrentAction} />

        <Grid item xs={12} sx={{ display: 'flex', justifyContent: 'flex-end' }}>
          <Button
            fullWidth
            startIcon={<Add />}
            variant='contained'
            color='secondary'
            onClick={() => {
              const act = valueToActionInput(currentAction)
              validationClient
                .query(query, {
                  input: act.dest,
                })
                .toPromise()
                .then((res) => {
                  if (res.error) {
                    console.error(res.error)
                    return
                  }

                  // clear the current action
                  setCurrentAction(
                    defaults(
                      types.find((t) => t.type === currentAction.destType) ||
                        types[0],
                    ),
                  )

                  props.onChange({
                    ...props.value,
                    actions: props.value.actions.concat(act),
                  })
                })
            }}
          >
            Add Action
          </Button>
        </Grid>
      </Grid>
    </FormContainer>
  )
}
