import React from 'react'
import {
  Grid,
  TextField,
  TextareaAutosize,
  makeStyles,
} from '@material-ui/core'
import { FormField } from '../../../forms'

const useStyles = makeStyles(theme => ({
  textarea: {
    font: 'inherit',
    fontFamily: theme.typography.fontFamily,
    color: 'currentColor',
    padding: 8,

    borderRadius: 4,
    border: '1px solid ' + theme.palette.primary['400'],
    '&:hover': {
      border: '2px solid black',
      padding: 7, // so content doesn't shift when border changes
    },
    '&:focus': {
      border: '2px solid ' + theme.palette.primary['500'],
      padding: 7, // so content doesn't shift when border changes
      outline: 0,
    },
  },
}))

export default props => {
  const classes = useStyles()

  return (
    <Grid container spacing={2}>
      <Grid item xs={12}>
        <FormField
          fullWidth
          label='Alert Summary'
          name='summary'
          required
          component={TextField}
        />
      </Grid>
      <Grid item xs={12}>
        <FormField
          fullWidth
          multiline
          rows={7}
          variant='outlined'
          placeholder='Alert Details'
          name='details'
          required
          component={TextareaAutosize}
          className={classes.textarea}
        />
      </Grid>
    </Grid>
  )
}
