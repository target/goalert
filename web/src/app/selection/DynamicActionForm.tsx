import React from 'react'
import {
  ActionInput,
  DestinationType,
  DestinationTypeInfo,
  ExprStringExpression,
} from '../../schema'
import { useDynamicActionTypes } from '../util/RequireConfig'
import { Grid, TextField } from '@mui/material'
import DestinationField from './DestinationField'
import { renderMenuItem } from './DisableableMenuItem'
import { HelperText } from '../forms'

export type StaticParams = Readonly<Record<string, string>>
export type DynamicParams = Readonly<Record<string, ExprStringExpression>>

export type Value = {
  destType: DestinationType
  staticParams: StaticParams
  dynamicParams: DynamicParams
}

export function valueToActionInput(value: Value | null): ActionInput {
  if (!value) {
    return { dest: { type: '', args: {} }, params: {} }
  }

  return {
    dest: {
      type: value.destType,
      args: value.staticParams,
    },
    params: value.dynamicParams,
  }
}

export function actionInputToValue(action: ActionInput): Value {
  return {
    destType: action.dest.type,
    staticParams: action.dest.args || {},
    dynamicParams: { ...action.params },
  }
}

export type DynamicActionErrors = {
  destTypeError?: string
  staticParamErrors?: Readonly<Record<string, string>>
  dynamicParamErrors?: Readonly<Record<string, string>>
}
export type DynamicActionFormProps = {
  value: Value | null
  onChange: (value: Value) => void

  disabled?: boolean

  destTypeError?: string
  staticParamErrors?: Readonly<Record<string, string>>
  dynamicParamErrors?: Readonly<Record<string, string>>

  disablePortal?: boolean
}

export function defaults(destTypeInfo: DestinationTypeInfo): Value {
  const staticParams = Object.fromEntries(
    destTypeInfo.requiredFields.map((f) => [f.fieldID, '']),
  )

  const dynamicParams = Object.fromEntries(
    destTypeInfo.dynamicParams.map((p) => [
      p.paramID,
      p.defaultValue || `req.body['${p.paramID}']`,
    ]),
  )

  return {
    destType: destTypeInfo.type,
    staticParams,
    dynamicParams,
  }
}

export default function DynamicActionForm(
  props: DynamicActionFormProps,
): React.ReactNode {
  const types = useDynamicActionTypes()
  const selectedDest = types.find((t) => t.type === props.value?.destType)

  return (
    <Grid item xs={12} container spacing={2}>
      <Grid item xs={12}>
        <TextField
          select
          fullWidth
          SelectProps={{ MenuProps: { disablePortal: props.disablePortal } }}
          value={selectedDest?.type || ''}
          label='Destination Type'
          name='dest.type'
          error={!!props.destTypeError}
          helperText={props.destTypeError}
          onChange={(e) => {
            const newType = types.find((t) => t.type === e.target.value)
            if (!newType) return
            props.onChange(defaults(newType))
          }}
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
      {props.value && (
        <Grid item xs={12}>
          <DestinationField
            value={props.value?.staticParams}
            onChange={(vals) => {
              if (!props.value) return
              props.onChange({
                ...props.value,
                staticParams: vals,
              })
            }}
            destType={props.value.destType}
            disabled={props.disabled}
            fieldErrors={props.staticParamErrors}
          />
        </Grid>
      )}
      {props.value &&
        (selectedDest?.dynamicParams || []).map((p) => {
          const fieldValue = props.value?.dynamicParams[p.paramID] || ''

          function handleChange(newValue: string): void {
            if (!props.value) return

            const newParams = {
              ...props.value.dynamicParams,
              [p.paramID]: newValue as ExprStringExpression,
            }

            props.onChange({ ...props.value, dynamicParams: newParams })
          }

          return (
            <Grid key={p.paramID} item xs={12}>
              <TextField
                fullWidth
                name={p.paramID}
                disabled={props.disabled || !selectedDest?.enabled}
                type='text'
                label={p.label + ' (Expr syntax)'}
                error={!!props.dynamicParamErrors?.[p.paramID]}
                helperText={
                  <HelperText
                    hint={p.hint}
                    hintURL={p.hintURL}
                    error={props.dynamicParamErrors?.[p.paramID]}
                  />
                }
                onChange={(e) => handleChange(e.target.value)}
                value={fieldValue}
              />
            </Grid>
          )
        })}
    </Grid>
  )
}
