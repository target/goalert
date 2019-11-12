import React from 'react'
import { Typography, Grid, Divider } from '@material-ui/core'
import { makeStyles } from '@material-ui/core/styles'
import { ServiceChip } from '../../../util/Chips'
import { FormField } from '../../../forms'
import Markdown from '../../../util/Markdown'

const useStyles = makeStyles(theme => ({
  divider: {
    backgroundColor: '#bdbdbd',
  },
  itemTitle: {
    paddingBottom: 0,
  },
  nudgeRight: {
    marginLeft: theme.spacing(1),
  },
}))

export function CreateAlertConfirm() {
  const classes = useStyles()

  const Item = ({ name, label, value, children }) => (
    <Grid item xs={12}>
      <Typography
        variant='subtitle1'
        component='h3'
        className={classes.itemTitle}
      >
        {label}
      </Typography>

      <Divider className={classes.divider} />

      {children ||
        (name === 'details' ? (
          <Markdown value={value} />
        ) : (
          <Typography
            variant='body1'
            component='p'
            className={classes.nudgeRight}
          >
            {value}
          </Typography>
        ))}
    </Grid>
  )

  return (
    <Grid container spacing={2}>
      <FormField
        name='summary'
        label='Summary'
        required
        render={props => <Item {...props} />}
      />
      <FormField
        name='details'
        label='Details'
        required
        render={props => <Item {...props} />}
      />

      <FormField
        label='Selected Services'
        name='serviceIDs'
        render={({ value, ...otherProps }) => (
          <Item {...otherProps} label={`Selected Services (${value.length})`}>
            {value.map(id => (
              <ServiceChip
                key={id}
                clickable={false}
                id={id}
                style={{ margin: 3 }}
                onClick={e => e.preventDefault()}
              />
            ))}
          </Item>
        )}
      />
    </Grid>
  )
}
