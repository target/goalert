import React, { useState } from 'react'
import _ from 'lodash'
import FormDialog from '../dialogs/FormDialog'
import { ActionInput, DynamicParamInput } from '../../schema'
import { FormContainer } from '../forms'
import { Grid, InputAdornment, TextField, Typography } from '@mui/material'
import { renderMenuItem } from '../selection/DisableableMenuItem'
import DestinationField from '../selection/DestinationField'
import { useDynamicActionTypes } from '../util/RequireConfig'

export default function RuleEditorActionDialog(props: {
  action: ActionInput
  onClose: (action: ActionInput | null) => void
}): JSX.Element {
  const [value, setValue] = useState(_.cloneDeep(props.action))
  const types = useDynamicActionTypes()
  const defaultParams = (typeName: string): DynamicParamInput[] =>
    (types.find((t) => t.type === typeName)?.dynamicParams || []).map((p) => ({
      paramID: p.paramID,
      expr: 'body.' + p.paramID,
    }))
  const selType = types.find((t) => t.type === value.dest.type)

  return (
    <FormDialog
      maxWidth='sm'
      title='Edit Action'
      onClose={() => props.onClose(null)}
      onSubmit={() => props.onClose(value)}
      form={
        <FormContainer>
          <Grid container spacing={2}>
            <Grid item xs={12}>
              <TextField
                fullWidth
                select
                label='Action Type'
                value={value.dest.type}
                onChange={(e) => {
                  setValue({
                    dest: { type: e.target.value as string, values: [] },
                    params: defaultParams(e.target.value as string),
                  })
                }}
              >
                {types.map((t) =>
                  renderMenuItem({
                    value: t.type,
                    label: t.name,
                    disabled: t.enabled === false,
                    disabledMessage: 'This action type is not enabled.',
                  }),
                )}
              </TextField>
            </Grid>
            <Grid item xs={12}>
              <DestinationField
                value={value.dest.values}
                destType={value.dest.type}
                onChange={(v) =>
                  setValue({ ...value, dest: { ...value.dest, values: v } })
                }
              />
            </Grid>
            {selType?.dynamicParams.map((p) => (
              <Grid item key={p.paramID} xs={12}>
                <TextField
                  fullWidth
                  multiline
                  InputProps={{
                    startAdornment: (
                      <InputAdornment position='start' sx={{ mb: '0.1em' }}>
                        {p.label}:
                      </InputAdornment>
                    ),
                    endAdornment: (
                      <InputAdornment position='end' sx={{ mb: '0.1em' }}>
                        {p.dataType}
                      </InputAdornment>
                    ),
                  }}
                  value={
                    value.params.find((x) => x.paramID === p.paramID)?.expr
                  }
                  onChange={(e) => {
                    const newParams = value.params.map((x) =>
                      x.paramID === p.paramID
                        ? { ...x, expr: e.target.value }
                        : x,
                    )
                    setValue({ ...value, params: newParams })
                  }}
                />
              </Grid>
            ))}
          </Grid>
        </FormContainer>
      }
    />
  )
}
