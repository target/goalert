import React from 'react'
import { DestinationType, FieldValueInput } from '../../schema'
import DestinationInputDirect from './DestinationInputDirect'
import { useDestinationType } from '../util/RequireConfig'
import DestinationSearchSelect from './DestinationSearchSelect'
import { Grid } from '@mui/material'
import { DestFieldValueError } from '../util/errtypes'

export type DestinationFieldProps = {
  value: FieldValueInput[]
  onChange?: (value: FieldValueInput[]) => void
  destType: DestinationType

  disabled?: boolean

  destFieldErrors?: DestFieldValueError[]
}

function capFirstLetter(s: string): string {
  if (s.length === 0) return s
  return s.charAt(0).toUpperCase() + s.slice(1)
}

export default function DestinationField(
  props: DestinationFieldProps,
): React.ReactNode {
  const dest = useDestinationType(props.destType)

  return (
    <Grid container spacing={2}>
      {dest.requiredFields.map((field) => {
        const fieldValue =
          (props.value || []).find((v) => v.fieldID === field.fieldID)?.value ||
          ''

        function handleChange(newValue: string): void {
          if (!props.onChange) return

          const newValues = (props.value || [])
            .filter((v) => v.fieldID !== field.fieldID)
            .concat({
              fieldID: field.fieldID,
              value: newValue,
            })

          props.onChange(newValues)
        }

        const fieldErrMsg = capFirstLetter(
          props.destFieldErrors?.find(
            (err) => err.extensions.fieldID === field.fieldID,
          )?.message || '',
        )

        if (field.supportsSearch)
          return (
            <Grid key={field.fieldID} item xs={12} sm={12} md={12}>
              <DestinationSearchSelect
                {...field}
                value={fieldValue}
                destType={props.destType}
                disabled={props.disabled || !dest.enabled}
                onChange={(val) => handleChange(val)}
                error={!!fieldErrMsg}
                hint={fieldErrMsg || field.hint}
                hintURL={fieldErrMsg ? '' : field.hintURL}
              />
            </Grid>
          )

        return (
          <Grid key={field.fieldID} item xs={12} sm={12} md={12}>
            <DestinationInputDirect
              {...field}
              value={fieldValue}
              destType={props.destType}
              disabled={props.disabled || !dest.enabled}
              onChange={(e) => handleChange(e.target.value)}
              error={!!fieldErrMsg}
              hint={fieldErrMsg || field.hint}
              hintURL={fieldErrMsg ? '' : field.hintURL}
            />
          </Grid>
        )
      })}
    </Grid>
  )
}
