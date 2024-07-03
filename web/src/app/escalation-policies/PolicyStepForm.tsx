import React, { useState, ReactNode, useEffect } from 'react'
import { FormContainer, FormField } from '../forms'
import Grid from '@mui/material/Grid'

import NumberField from '../util/NumberField'
import { DestinationInput, StringMap } from '../../schema'
import DestinationInputChip from '../util/DestinationInputChip'
import { TextField, Typography } from '@mui/material'
import { renderMenuItem } from '../selection/DisableableMenuItem'
import DestinationField from '../selection/DestinationField'
import { useEPTargetTypes } from '../util/RequireConfig'
import { gql, useClient, CombinedError } from 'urql'
import DialogContentError from '../dialogs/components/DialogContentError'
import makeStyles from '@mui/styles/makeStyles'
import { useErrorConsumer } from '../util/ErrorConsumer'
import { Add } from '../icons'
import LoadingButton from '../loading/components/LoadingButton'

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

export type PolicyStepFormProps = {
  value: FormValue
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

export default function PolicyStepForm(props: PolicyStepFormProps): ReactNode {
  const types = useEPTargetTypes()
  const classes = useStyles()
  const [destType, setDestType] = useState(types[0].type)
  const [args, setArgs] = useState<StringMap>({})
  const validationClient = useClient()
  const [err, setErr] = useState<CombinedError | null>(null)
  const errs = useErrorConsumer(err)
  const [validating, setValidating] = useState(false)

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

  function renderErrors(otherErrs: readonly string[]): React.JSX.Element[] {
    return otherErrs.map((err, idx) => (
      <DialogContentError
        error={err}
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
    >
      <Grid container spacing={2}>
        <Grid container spacing={1} item xs={12} sx={{ p: 1 }}>
          {props.value.actions.map((a) => (
            <Grid item key={JSON.stringify(a.values)}>
              <DestinationInputChip
                value={a}
                onDelete={props.disabled ? undefined : () => handleDelete(a)}
              />
            </Grid>
          ))}
          {props.value.actions.length === 0 && (
            <Grid item xs={12}>
              <Typography variant='body2' color='textSecondary'>
                No destinations
              </Typography>
            </Grid>
          )}
        </Grid>
        <Grid item xs={12}>
          <TextField
            select
            fullWidth
            disabled={props.disabled || validating}
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
            value={args}
            disabled={props.disabled || validating}
            onChange={(newValue: StringMap) => {
              setErr(null)
              setArgs(newValue)
            }}
            fieldErrors={errs.getErrorMap(
              /destinationDisplayInfo\.input(\.args)?/,
            )}
          />
        </Grid>
        <Grid container item xs={12} justifyContent='flex-end'>
          {errs.hasErrors() && renderErrors(errs.remaining())}
          <LoadingButton
            variant='contained'
            color='secondary'
            fullWidth
            style={{ width: '100%' }}
            loading={validating}
            disabled={props.disabled}
            startIcon={<Add />}
            noSubmit
            onClick={() => {
              if (!props.onChange) return
              setValidating(true)
              validationClient
                .query(query, {
                  input: {
                    type: destType,
                    args,
                  },
                })
                .toPromise()
                .then((res) => {
                  setValidating(false)
                  if (res.error) {
                    setErr(res.error)
                    return
                  }
                  setArgs({})
                  props.onChange({
                    ...props.value,
                    actions: props.value.actions.concat({
                      type: destType,
                      args,
                    }),
                  })
                })
            }}
          >
            Add Destination
          </LoadingButton>
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
