import React, { ReactNode } from 'react'
import { Typography, Grid, Divider } from '@mui/material'
import { ServiceChip } from '../../../util/ServiceChip'
import { FormField } from '../../../forms'
import Markdown from '../../../util/Markdown'

type FieldProps = {
  children: ReactNode
  label: string
}

function Field(props: FieldProps): React.JSX.Element {
  return (
    <Grid size={12}>
      <Typography
        variant='subtitle1'
        component='h3'
        sx={{ pb: 0 }}
      >
        {props.label}
      </Typography>

      <Divider />

      <div style={{ marginTop: '0.5em' }}>{props.children}</div>
    </Grid>
  )
}

export function CreateAlertConfirm(): React.JSX.Element {
  return (
    <Grid container spacing={2}>
      <FormField
        name='summary'
        required
        render={(p: { value: string }) => (
          <Field label='Summary'>
            <Typography variant='body1' component='p'>
              {p.value}
            </Typography>
          </Field>
        )}
      />
      <FormField
        name='details'
        render={(p: { value: string }) => (
          <Field label='Details'>
            <Typography variant='body1' component='div'>
              <Markdown value={p.value} />
            </Typography>
          </Field>
        )}
      />
      <FormField
        name='serviceIDs'
        render={(p: { value: string[] }) => (
          <Field label={`Selected Services (${p.value.length})`}>
            {p.value.map((id: string) => (
              <ServiceChip
                key={id}
                clickable={false}
                id={id}
                style={{ margin: 3 }}
                onClick={(e) => e.preventDefault()}
              />
            ))}
          </Field>
        )}
      />
    </Grid>
  )
}
