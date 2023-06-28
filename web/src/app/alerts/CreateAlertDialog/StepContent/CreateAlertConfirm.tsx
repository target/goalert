import React from 'react'
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

export function CreateAlertConfirm(): JSX.Element {
  const classes = useStyles()

  const renderItem = ({
    name,
    label,
    value,
    children,
  }: {
    name: string
    label: string
    value: string
    children: JSX.Element
  }): JSX.Element => (
    <Grid item xs={12}>
      <Typography
        variant='subtitle1'
        component='h3'
        className={classes.itemTitle}
      >
        {label}
      </Typography>

      <Divider />

      <div className={classes.itemContent}>
        {children ||
          (name === 'details' ? (
            <Typography variant='body1' component='div'>
              <Markdown value={value} />
            </Typography>
          ) : (
            <Typography variant='body1' component='p'>
              {value}
            </Typography>
          ))}
      </div>
    </Grid>
  )

  return (
    <Grid container spacing={2}>
      <FormField name='summary' label='Summary' required render={renderItem} />
      <FormField name='details' label='Details' render={renderItem} />

      <FormField
        label='Selected Services'
        name='serviceIDs'
        render={({ value, ...otherProps }) =>
          renderItem({
            ...otherProps,
            label: `Selected Services (${value.length})`,
            children: value.map((id: string) => (
              <ServiceChip
                key={id}
                clickable={false}
                id={id}
                style={{ margin: 3 }}
                onClick={(e) => e.preventDefault()}
              />
            )),
          })
        }
      />
    </Grid>
  )
}
