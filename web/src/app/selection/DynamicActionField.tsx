import React from 'react'
import {
  ActionInput,
  DestinationType,
  DestinationTypeInfo,
  ExprStringExpression,
  FieldValueInput,
} from '../../schema'
import {
  useDestinationType,
  useDynamicActionTypes,
} from '../util/RequireConfig'
import { Grid, TextField } from '@mui/material'
import { DestFieldValueError } from '../util/errtypes'
import DestinationField from './DestinationField'
import { renderMenuItem } from './DisableableMenuItem'
import AppLink from '../util/AppLink'

export type StaticParams = Map<string, string>
export type DynamicParams = Map<string, ExprStringExpression>

export type Value = {
  destType: DestinationType
  staticParams: Map<string, string>
  dynamicParams: Map<string, ExprStringExpression>
}

export function valueToActionInput(value: Value): ActionInput {
  return {
    dest: {
      type: value.destType,
      values: Array.from(value.staticParams.entries()).map(([k, v]) => ({
        fieldID: k,
        value: v,
      })),
    },
    params: Array.from(value.dynamicParams.entries()).map(([k, v]) => ({
      paramID: k,
      expr: v,
    })),
  }
}

export function actionInputToValue(action: ActionInput): Value {
  return {
    destType: action.dest.type,
    staticParams: new Map(action.dest.values.map((v) => [v.fieldID, v.value])),
    dynamicParams: new Map(action.params.map((p) => [p.paramID, p.expr])),
  }
}

export function staticToDestField(
  staticParams: StaticParams,
): FieldValueInput[] {
  return Array.from(staticParams.entries()).map(([k, v]) => ({
    fieldID: k,
    value: v,
  }))
}
export function destFieldToStatic(destFields: FieldValueInput[]): StaticParams {
  return new Map(destFields.map((f) => [f.fieldID, f.value]))
}

export type DynamicActionFieldProps = {
  value: Value
  onChange: (value: Value) => void

  disabled?: boolean

  destFieldErrors?: DestFieldValueError[]
}

export function defaults(destTypeInfo: DestinationTypeInfo): Value {
  const staticParams = new Map()
  destTypeInfo.requiredFields.forEach((f) => {
    staticParams.set(f.fieldID, '')
  })
  const dynamicParams = new Map()
  destTypeInfo.dynamicParams.forEach((p) => {
    dynamicParams.set(p.paramID, `req.body.${p.paramID}`)
  })

  return {
    destType: destTypeInfo.type,
    staticParams,
    dynamicParams,
  }
}

// DynamicActionField renders a select dropdown to choose what type of Destination
//  to use, followed by all the fields that Destination requires.
//
// e.g. When Destination Type = Slack is selected, the form may render
// a DestinationSearchSelect to allow the user to find which channel or user.
export default function DynamicActionField(
  props: DynamicActionFieldProps,
): React.ReactNode {
  const types = useDynamicActionTypes()
  const dest = useDestinationType(props.value.destType)

  return (
    <Grid container spacing={2} item xs={12}>
      {/* Choose your destination */}
      <Grid item xs={12}>
        <TextField
          select
          fullWidth
          value={props.value.destType}
          label='Destination Type'
          name='dest.type'
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

      {/* Static fields for things like the address of the message */}
      <Grid item xs={12}>
        <DestinationField
          value={staticToDestField(props.value.staticParams)}
          onChange={(vals) =>
            props.onChange({
              ...props.value,
              staticParams: destFieldToStatic(vals),
            })
          }
          destType={props.value.destType}
          disabled={props.disabled}
          destFieldErrors={props.destFieldErrors}
        />
      </Grid>

      {/* Dynamic fields are used for things like the body of the message */}
      {(dest.dynamicParams || []).map((p) => {
        const fieldValue = props.value.dynamicParams.get(p.paramID) || ''

        function handleChange(newValue: string): void {
          const newParams = new Map(props.value.dynamicParams)
          newParams.set(p.paramID, newValue)

          props.onChange({ ...props.value, dynamicParams: newParams })
        }

        return (
          <Grid key={p.paramID} item xs={12}>
            <TextField
              fullWidth
              name={p.paramID}
              disabled={props.disabled || !dest.enabled}
              type='text'
              label={p.label + ' (Expr syntax)'}
              helperText={
                p.hintURL ? (
                  <AppLink newTab to={p.hintURL}>
                    {p.hint}
                  </AppLink>
                ) : (
                  p.hint
                )
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
