import React, { ReactNode } from 'react'
import { Typography, Grid, Divider } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import { ServiceChip } from '../../../util/Chips'
import { FormField } from '../../../forms'
import Markdown from '../../../util/Markdown'

const useStyles = makeStyles({
  itemContent: {
    marginTop: '0.5em',
  },
  itemTitle: {
    paddingBottom: 0,
  },
})

type FieldProps = {
  children: ReactNode
  label: string
}

function Field(props: FieldProps): JSX.Element {
  const classes = useStyles()
  return (
    <Grid item xs={12}>
      <Typography
        variant='subtitle1'
        component='h3'
        className={classes.itemTitle}
      >
        {props.label}
      </Typography>

      <Divider />

      <div className={classes.itemContent}>{props.children}</div>
    </Grid>
  )
}

export function CreateAlertConfirm(): JSX.Element {
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
