import React from 'react'
import { Paper, Typography } from '@material-ui/core'
import { makeStyles } from '@material-ui/core/styles'
import { ServiceChip } from '../../../util/Chips'

const useStyles = makeStyles(theme => ({
  nudgeRight: {
    marginLeft: theme.spacing(1.3),
  },
}))

export default props => {
  const { formFields } = props

  const classes = useStyles()

  return (
    <Paper elevation={0}>
      <Typography variant='subtitle1' component='h3'>
        Summary
      </Typography>
      <Typography variant='body1' component='p' className={classes.nudgeRight}>
        {formFields.summary}
      </Typography>
      <Typography variant='subtitle1' component='h3'>
        Details
      </Typography>
      <Typography variant='body1' component='p' className={classes.nudgeRight}>
        {formFields.details}
      </Typography>
      <Typography variant='subtitle1' component='h3'>
        {`Selected Services (${formFields.selectedServices.length})`}
      </Typography>

      {formFields.selectedServices.length > 0 && (
        <span>
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
        </span>
      )}
    </Paper>
  )
}
