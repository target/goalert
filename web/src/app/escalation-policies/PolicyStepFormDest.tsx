import React, { useState, ReactNode, useEffect } from 'react'
import { FormContainer, FormField } from '../forms'
import Grid from '@mui/material/Grid'

import NumberField from '../util/NumberField'
import { DestinationInput, FieldValueInput } from '../../schema'
import DestinationInputChip from '../util/DestinationInputChip'
import { Button, Divider, TextField, Typography } from '@mui/material'
import { renderMenuItem } from '../selection/DisableableMenuItem'
import DestinationField from '../selection/DestinationField'
import { useEPTargetTypes } from '../util/RequireConfig'
import { gql, useClient, CombinedError } from 'urql'
import {
  DestFieldValueError,
  KnownError,
  isDestFieldError,
} from '../util/errtypes'
import { splitErrorsByPath } from '../util/errutil'
import DialogContentError from '../dialogs/components/DialogContentError'
import makeStyles from '@mui/styles/makeStyles'

const useStyles = makeStyles(() => {
  return {
    errorContainer: {
      flexGrow: 0,
      overflowY: 'visible',
    },
  }
})

export type FormValue = {
  delayMinutes: number
  actions: DestinationInput[]
}

export type PolicyStepFormDestProps = {
  value: FormValue
  errors?: (KnownError | DestFieldValueError)[]
  disabled?: boolean
  onChange: (value: FormValue) => void
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
  const classes = useStyles()
  const [destType, setDestType] = useState(types[0].type)
  const [values, setValues] = useState<FieldValueInput[]>([])
  const validationClient = useClient()
  const [err, setErr] = useState<CombinedError | null>(null)
  const [destErrors, otherErrs] = splitErrorsByPath(err || props.errors, [
    'destinationDisplayInfo.input',
  ])

  useEffect(() => {
    setErr(null)
  }, [props.value])

  function handleDelete(a: DestinationInput): void {
    if (!props.onChange) return
    props.onChange({
      ...props.value,
      actions: props.value.actions.filter((b) => a !== b),
    })
  }

  function renderErrors(): React.JSX.Element[] {
    return otherErrs.map((err, idx) => (
      <DialogContentError
        error={err.message || err}
        key={idx}
        noPadding
        className={classes.errorContainer}
      />
    ))
  }

  return (
    <FormContainer
      value={props.value}
      onChange={(newValue: FormValue) => {
        if (!props.onChange) return
        props.onChange(newValue)
      }}
      optionalLabels
      errors={props.errors}
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
            disabled={props.disabled}
            onChange={(newValue: FieldValueInput[]) => {
              setErr(null)
              setValues(newValue)
            }}
            destFieldErrors={destErrors.filter(isDestFieldError)}
          />
        </Grid>
        <Grid container item xs={12} justifyContent='flex-end'>
          {otherErrs && renderErrors()}
          <Button
            variant='contained'
            color='secondary'
            onClick={() => {
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
                    setErr(res.error)
                    return
                  }
                  setValues([])
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
            Add Destination
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
