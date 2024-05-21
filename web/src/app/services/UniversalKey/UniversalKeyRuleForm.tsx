import React, { useState } from 'react'
import { FormContainer, FormField } from '../../forms'
import { Button, Grid, TextField, Typography } from '@mui/material'
import { ActionInput, KeyRuleInput } from '../../../schema'
import { useDynamicActionTypes } from '../../util/RequireConfig'
import DestinationInputChip from '../../util/DestinationInputChip'
import { gql, useClient } from 'urql'
import DynamicActionField, {
  Value as ActionValue,
  valueToActionInput,
} from '../../selection/DynamicActionField'

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

  const [currentAction, setCurrentAction] = useState<ActionValue>({
    destType: types[0].type,
    staticParams: new Map(),
    dynamicParams: new Map(),
  })

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
            label='Expr'
            name='conditionExpr'
            required
            multiline
            rows={3}
          />
        </Grid>
        <Grid item xs={12}>
          <Typography variant='h6' color='textPrimary'>
            Actions
          </Typography>
        </Grid>
        <Grid item xs={12}>
          {props.value.actions.map((a, idx) => (
            <DestinationInputChip
              key={idx}
              value={a.dest}
              onDelete={() => handleDelete(a)}
            />
          ))}
          {props.value.actions.length === 0 && (
            <Typography variant='body2' color='textSecondary'>
              No actions
            </Typography>
          )}
        </Grid>
        <DynamicActionField value={currentAction} onChange={setCurrentAction} />

        <Grid container item xs={12} justifyContent='flex-end'>
          <Button
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
                    return
                  }

                  // clear the current action
                  setCurrentAction({
                    destType: currentAction.destType,
                    staticParams: new Map(),
                    dynamicParams: new Map(),
                  })

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
