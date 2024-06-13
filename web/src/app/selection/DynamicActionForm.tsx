import React from 'react'
import {
  ActionInput,
  DestinationType,
  DestinationTypeInfo,
  ExprStringExpression,
  FieldValueInput,
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
    return { dest: { type: '', values: [] }, params: {} }
  }

  return {
    dest: {
      type: value.destType,
      values: Object.entries(value.staticParams).map(([fieldID, value]) => ({
        fieldID,
        value,
      })),
    },
    params: value.dynamicParams,
  }
}

export function actionInputToValue(action: ActionInput): Value {
  return {
    destType: action.dest.type,
    staticParams: Object.fromEntries(
      action.dest.values.map((v) => [v.fieldID, v.value]),
    ),
    dynamicParams: { ...action.params },
  }
}

export function staticToDestField(
  staticParams?: StaticParams,
): FieldValueInput[] {
  if (!staticParams) return []
  return Object.entries(staticParams).map(([fieldID, value]) => ({
    fieldID,
    value,
  }))
}

export function destFieldToStatic(destFields: FieldValueInput[]): StaticParams {
  return Object.fromEntries(destFields.map((f) => [f.fieldID, f.value]))
}

export type DynamicActionErrors = {}
export type DynamicActionFormProps = {
  value: Value | null
  onChange: (value: Value) => void

  disabled?: boolean

  destTypeError?: string
  staticParamErrors?: Readonly<Record<string, string>>
  dynamicParamErrors?: Readonly<Record<string, string>>
}

export function defaults(destTypeInfo: DestinationTypeInfo): Value {
  const staticParams = Object.fromEntries(
    destTypeInfo.requiredFields.map((f) => [f.fieldID, '']),
  )

  const dynamicParams = Object.fromEntries(
    destTypeInfo.dynamicParams.map((p) => [p.paramID, `req.body.${p.paramID}`]),
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
    <Grid container spacing={2} item xs={12}>
      <Grid item xs={12}>
        <TextField
          select
          fullWidth
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
            value={staticToDestField(props.value?.staticParams)}
            onChange={(vals) => {
              if (!props.value) return
              props.onChange({
                ...props.value,
                staticParams: destFieldToStatic(vals),
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
