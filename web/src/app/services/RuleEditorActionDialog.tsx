import React, { useState } from 'react'

import FormDialog from '../dialogs/FormDialog'
import { DestinationInput, FieldValueInput } from '../../schema'
import { FormContainer, FormField } from '../forms'
import { Grid, Input, InputAdornment, TextField } from '@mui/material'

export type Action = {
  dest: DestinationInput

  params: FieldValueInput[]
}

function ActionForm() {
  return (
    <div>
      <h1>Action Form</h1>
    </div>
  )
}

export default function RuleEditorActionDialog(props: {
  onClose: (expr: string | null) => void
}): JSX.Element {
  const [value, setValue] = useState<string>(props.expr)

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
                InputProps={{
                  startAdornment: (
                    <InputAdornment position='start' sx={{ mb: '0.1em' }}>
                      Summary:
                    </InputAdornment>
                  ),
                  endAdornment: (
                    <InputAdornment position='end' sx={{ mb: '0.1em' }}>
                      string
                    </InputAdornment>
                  ),
                }}
                value='body.title'
              />
            </Grid>
            <Grid item xs={12}>
              <TextField
                fullWidth
                InputProps={{
                  startAdornment: (
                    <InputAdornment position='start' sx={{ mb: '0.1em' }}>
                      Details:
                    </InputAdornment>
                  ),
                  endAdornment: (
                    <InputAdornment position='end' sx={{ mb: '0.1em' }}>
                      string
                    </InputAdornment>
                  ),
                }}
                value='body.details'
              />
            </Grid>
            <Grid item xs={12}>
              <TextField
                fullWidth
                InputProps={{
                  startAdornment: (
                    <InputAdornment position='start' sx={{ mb: '0.1em' }}>
                      Dedup:
                    </InputAdornment>
                  ),
                  endAdornment: (
                    <InputAdornment position='end' sx={{ mb: '0.1em' }}>
                      string
                    </InputAdornment>
                  ),
                }}
                value='body.title + body.details'
              />
            </Grid>
            <Grid item xs={12}>
              <TextField
                fullWidth
                InputProps={{
                  startAdornment: (
                    <InputAdornment position='start' sx={{ mb: '0.1em' }}>
                      Close Alert?:
                    </InputAdornment>
                  ),
                  endAdornment: (
                    <InputAdornment position='end' sx={{ mb: '0.1em' }}>
                      bool
                    </InputAdornment>
                  ),
                }}
                value='body.state == "firing"'
              />
            </Grid>
          </Grid>
        </FormContainer>
      }
    />
  )
}
