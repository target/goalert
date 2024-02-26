import React, { useState, ReactNode } from 'react'
import { FormContainer, FormField } from '../forms'
import Grid from '@mui/material/Grid'

import NumberField from '../util/NumberField'
import { DestinationInput, FieldValueInput } from '../../schema'
import DestinationInputChip from '../util/DestinationInputChip'
import { Button, TextField, Typography } from '@mui/material'
import { renderMenuItem } from '../selection/DisableableMenuItem'
import DestinationField from '../selection/DestinationField'
import { useEPTargetTypes } from '../util/RequireConfig'
import { gql, useClient, CombinedError } from 'urql'
import {
  DestFieldValueError,
  KnownError,
  isInputFieldError,
} from '../util/errtypes'

export type FormValue = {
  delayMinutes: number
  actions: DestinationInput[]
}

export type PolicyStepFormDestProps = {
  value: FormValue
  errors?: (KnownError | DestFieldValueError)[]
  disabled?: boolean
  onChange: (value: FormValue) => void
  setErr: (err: CombinedError | null) => void
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

export default function PolicyStepFormDest(
  props: PolicyStepFormDestProps,
): ReactNode {
  const types = useEPTargetTypes()

  const [destType, setDestType] = useState(types[0].type)
  const [values, setValues] = useState<FieldValueInput[]>([])
  const validationClient = useClient()

  function handleDelete(a: DestinationInput): void {
    if (!props.onChange) return

    props.onChange({
      ...props.value,
      actions: props.value.actions.filter((b) => a !== b),
    })
  }

  return (
    <FormContainer
      value={props.value}
      onChange={(newValue: FormValue) => {
        if (!props.onChange) return

        props.onChange(newValue)
      }}
      errors={props.errors?.filter(isInputFieldError).map((e) => {
        let field = e.path[e.path.length - 1].toString()
        if (field === 'type') field = 'dest.type'
        return {
          // need to convert to FormContainer's error format
          message: e.message,
          field,
        }
      })}
    >
      <Grid container spacing={2}>
        <Grid item xs={12}>
          {props.value.actions.map((a, idx) => (
            <DestinationInputChip
              key={idx}
              value={a}
              onDelete={props.disabled ? undefined : () => handleDelete(a)}
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
            disabled={props.disabled}
            value={destType}
            onChange={(e) => setDestType(e.target.value)}
          >
            {types.map((t) =>
              renderMenuItem({
                label: t.name,
                value: t.type,
                disabled: !t.enabled,
                disabledMessage: t.enabled ? '' : t.disabledMessage,
              }),
            )}
          </TextField>
        </Grid>
        <Grid item xs={12}>
          <DestinationField
            destType={destType}
            value={values}
            disabled={props.disabled}
            onChange={(newValue: FieldValueInput[]) => setValues(newValue)}
          />
        </Grid>
        <Grid container item xs={12} justifyContent='flex-end'>
          <Button
            variant='contained'
            onClick={() => {
              // setValues([])
              if (!props.onChange) return

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
                    props.setErr(res.error)
                    return
                  }

                  props.onChange({
                    ...props.value,
                    actions: props.value.actions.concat({
                      type: destType,
                      values,
                    }),
                  })
                })
            }}
          >
            Add Action
          </Button>
        </Grid>
        <Grid item xs={12}>
          <FormField
            component={NumberField}
            disabled={props.disabled}
            fullWidth
            label='Delay (minutes)'
            name='delayMinutes'
            required
            min={1}
            max={9000}
            hint={
              props.value.delayMinutes === 0
                ? 'This will cause the step to immediately escalate'
                : `This will cause the step to escalate after ${props.value.delayMinutes}m`
            }
          />
        </Grid>
      </Grid>
    </FormContainer>
  )
}
