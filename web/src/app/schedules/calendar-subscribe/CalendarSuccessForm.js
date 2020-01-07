import {
  Button,
  FormHelperText,
  Grid,
  TextField,
  Typography,
} from '@material-ui/core'
import { PropTypes as p } from 'prop-types'
import React from 'react'

export default function CalenderSuccessForm(props) {
  const style = { display: 'flex', justifyContent: 'center' }
  return (
    <Grid container spacing={2}>
      <Grid item xs={12} style={style}>
        <Typography>
          Your subscription has been created! You can manage your subscriptions
          from your profile at anytime.
        </Typography>
      </Grid>
      <Grid item xs={12} style={style}>
        <Button
          color='primary'
          variant='contained'
          href={'webcal://' + props.url}
          style={{ marginLeft: '0.5em' }}
        >
          Subscribe
        </Button>
      </Grid>
      <Grid item xs={12}>
        <TextField
          value={props.url}
          onChange={() => {}}
          style={{ width: '100%' }}
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
