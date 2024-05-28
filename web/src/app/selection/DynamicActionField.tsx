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
import AppLink from '../util/AppLink'

export type StaticParams = Readonly<Record<string, string>>
export type DynamicParams = Readonly<Record<string, ExprStringExpression>>

export type Value = {
  destType: DestinationType
  staticParams: StaticParams
  dynamicParams: DynamicParams
}

export function valueToActionInput(value: Value): ActionInput {
  return {
    dest: {
      type: value.destType,
      values: Object.entries(value.staticParams).map(([fieldID, value]) => ({
        fieldID,
        value,
      })),
    },
    params: Object.entries(value.dynamicParams).map(([paramID, expr]) => ({
      paramID,
      expr,
    })),
  }
}

export function actionInputToValue(action: ActionInput): Value {
  return {
    destType: action.dest.type,
    staticParams: Object.fromEntries(
      action.dest.values.map((v) => [v.fieldID, v.value]),
    ),
    dynamicParams: Object.fromEntries(
      action.params.map((p) => [p.paramID, p.expr]),
    ),
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

interface BaseError {
  message: string
}

interface InputError extends BaseError {
  section: 'input'
  inputID: 'dest-type'
}
interface StaticFieldError extends BaseError {
  section: 'static-params'
  fieldID: string
}

interface DynamicParamError extends BaseError {
  section: 'dynamic-params'
  paramID: string
}

export type FormError = InputError | StaticFieldError | DynamicParamError

export type DynamicActionFieldProps = {
  value: Value | null
  onChange: (value: Value) => void

  disabled?: boolean

  errors?: Array<FormError>
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

function isStatic(err: BaseError): err is StaticFieldError {
  return (err as StaticFieldError).section === 'static-params'
}
function isInput(err: BaseError): err is InputError {
  return (err as InputError).section === 'input'
}
function isDynamic(err: BaseError): err is DynamicParamError {
  return (err as DynamicParamError).section === 'dynamic-params'
}

export default function DynamicActionField(
  props: DynamicActionFieldProps,
): React.ReactNode {
  const types = useDynamicActionTypes()
  const selectedDest = types.find((t) => t.type === props.value?.destType)

  const typeError = props.errors?.find(isInput)
  const dynamicErrorMap = new Map(
    props.errors?.filter(isDynamic).map((e) => [e.paramID, e.message]),
  )

  return (
    <Grid container spacing={2} item xs={12}>
      <Grid item xs={12}>
        <TextField
          select
          fullWidth
          value={selectedDest?.type || ''}
          label='Destination Type'
          name='dest.type'
          error={!!typeError}
          helperText={typeError?.message}
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
            fieldErrors={props.errors?.filter(isStatic)}
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
                error={!!dynamicErrorMap.get(p.paramID)}
                helperText={
                  dynamicErrorMap.get(p.paramID) ||
                  (p.hintURL ? (
                    <AppLink newTab to={p.hintURL}>
                      {p.hint}
                    </AppLink>
                  ) : (
                    p.hint
                  ))
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
