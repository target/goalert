import React from 'react'
import { Paper, Typography } from '@material-ui/core'
import { makeStyles } from '@material-ui/core/styles'
import { ServiceChip } from '../../../util/Chips'

const useStyles = makeStyles(theme => ({
  hrUnderline: {
    marginTop: 0,
    marginBottom: 3,
  },
  item: { marginBottom: theme.spacing(2) },

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
    <div className={classes.item}>
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
    </div>
  )

  return (
    <Paper elevation={0}>
      <Item title={'Summary'} description={formFields.summary} />
      <Item title={'Details'} description={formFields.details} />

      <Item title={`Selected Services (${formFields.selectedServices.length})`}>
        <Paper elevation={0}>
          {formFields.selectedServices.map((id, key) => (
            <ServiceChip
              key={key}
              clickable={false}
              id={id}
              style={{ margin: 3 }}
              onClick={e => e.preventDefault()}
            />
          ))}
        </Paper>
      </Item>
    </Paper>
  )
}
