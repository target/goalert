import React from 'react'
import { Typography, Grid, Divider } from '@material-ui/core'
import { makeStyles } from '@material-ui/core/styles'
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
  markdown: {
    '& p': {
      marginBlockStart: 0,
      marginBlockEnd: 0,
      whiteSpace: 'pre-wrap',
    },
  },
})

export function CreateAlertConfirm() {
  const classes = useStyles()

  const renderItem = ({ name, label, value, children }) => (
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
            <Typography
              variant='body1'
              className={classes.markdown}
              component={Markdown}
              value={value}
            />
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
            children: value.map(id => (
              <ServiceChip
                key={id}
                clickable={false}
                id={id}
                style={{ margin: 3 }}
                onClick={e => e.preventDefault()}
              />
            )),
          })
        }
      />
    </Grid>
  )
}
