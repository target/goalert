import React from 'react'
import { PropTypes as p } from 'prop-types'
import {
  Button,
  FormHelperText,
  Grid,
  TextField,
  Typography,
  makeStyles,
} from '@material-ui/core'

const useStyles = makeStyles({
  caption: {
    width: '100%',
  },
  item: {
    display: 'flex',
    justifyContent: 'center',
  },
})

export default function CalenderSuccessForm(props) {
  const classes = useStyles()
  return (
    <Grid container spacing={2}>
      <Grid item xs={12} className={classes.item}>
        <Typography>
          Your subscription has been created! You can manage your subscriptions
          from your profile at anytime.
        </Typography>
      </Grid>
      <Grid item xs={12} className={classes.item}>
        <Button
          color='primary'
          variant='contained'
          href={'webcal://' + props.url}
        >
          Subscribe
        </Button>
      </Grid>
      <Grid item xs={12}>
        <TextField
          value={props.url}
          onChange={() => {}}
          className={classes.caption}
        />
        <FormHelperText>
          Some applications require you copy and paste the URL directly
        </FormHelperText>
      </Grid>
    </Grid>
  )
}

CalenderSuccessForm.propTypes = {
  url: p.string.isRequired,
}
