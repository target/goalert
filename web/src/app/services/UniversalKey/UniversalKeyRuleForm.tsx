import React, { ReactElement, useState } from 'react'
import { FormContainer, FormField } from '../../forms'
import { Button, Grid, TextField, Typography } from '@mui/material'
import { ActionInput, FieldValueInput, KeyRuleInput } from '../../../schema'
import { renderMenuItem } from '../../selection/DisableableMenuItem'
import { useEPTargetTypes } from '../../util/RequireConfig'
import DestinationField from '../../selection/DestinationField'
import DestinationInputChip from '../../util/DestinationInputChip'
import { gql, useClient } from 'urql'

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
  const types = useEPTargetTypes()
  const [destType, setDestType] = useState(types[0].type)
  const [values, setValues] = useState<FieldValueInput[]>([])
  const validationClient = useClient()

  function handleDelete(a: ActionInput): void {
    if (!props.onChange) return
    props.onChange({
      ...props.value,
      actions: props.value.actions.filter((b) => a !== b),
    })
  }

  function renderAction(): ReactElement {
    return (
      <React.Fragment>
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
        <Grid item xs={12}>
          <TextField
            select
            fullWidth
            value={destType}
            label='Destination Type'
            name='dest.type'
            onChange={(e) => setDestType(e.target.value)}
          >
            {types.map((t) =>
              renderMenuItem({
                label: t.name,
                value: t.type,
                disabled: !t.enabled,
                disabledMessage: t.enabled ? '' : 'Disabled by administrator.',
              }),
            )}
          </TextField>
        </Grid>
        <Grid item xs={12}>
          <DestinationField
            destType={destType}
            value={values}
            onChange={(newValue: FieldValueInput[]) => {
              setValues(newValue)
            }}
          />
        </Grid>

        {/* TODO: add dynamic action params */}

        <Grid container item xs={12} justifyContent='flex-end'>
          <Button
            variant='contained'
            color='secondary'
            onClick={() => {
              validationClient
                .query(query, {
                  input: {
                    type: destType,
                    values,
                  },
                })
                .toPromise()
                .then((res) => {
                  if (res.error) {
                    return
                  }
                  setValues([])
                  props.onChange({
                    ...props.value,
                    actions: props.value.actions.concat({
                      dest: {
                        type: destType,
                        values,
                      },
                      params: [],
                    }),
                  })
                })
            }}
          >
            Add Destination
          </Button>
        </Grid>
      </React.Fragment>
    )
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
        {renderAction()}
      </Grid>
    </FormContainer>
  )
}
