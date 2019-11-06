import React from 'react'
import { Typography, Grid } from '@material-ui/core'
import { makeStyles } from '@material-ui/core/styles'
import { ServiceChip } from '../../../util/Chips'

const useStyles = makeStyles(theme => ({
  hrUnderline: {
    marginTop: 0,
    marginBottom: 3,
  },
  itemTitle: {
    paddingBottom: 0,
  },
  nudgeRight: {
    marginLeft: theme.spacing(1),
  },
}))

export default props => {
  const { formFields } = props

  const classes = useStyles()

  const Item = props => (
    <Grid item xs={12}>
      <Typography
        variant='subtitle1'
        component='h3'
        className={classes.itemTitle}
      >
        {props.title}
      </Typography>
      <hr className={classes.hrUnderline} />

      {props.description && (
        <Typography
          variant='body1'
          component='p'
          className={classes.nudgeRight}
        >
          {props.description}
        </Typography>
      )}

      {props.children}
    </Grid>
  )

  return (
    <Grid container spacing={2}>
      <Item title={'Summary'} description={formFields.Summary} />
      <Item title={'Details'} description={formFields.Details} />

      <Item title={`Selected Services (${formFields.selectedServices.length})`}>
        {formFields.selectedServices.map((id, key) => (
          <ServiceChip
            key={key}
            clickable={false}
            id={id}
            style={{ margin: 3 }}
            onClick={e => e.preventDefault()}
          />
        ))}
      </Item>
    </Grid>
  )
}
