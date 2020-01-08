import React from 'react'
import { PropTypes as p } from 'prop-types'
import {
  Button,
  FormHelperText,
  Grid,
  Typography,
  makeStyles,
} from '@material-ui/core'
import CopyText from '../../util/CopyText'

const useStyles = makeStyles({
  caption: {
    width: '100%',
  },
  subscribeButtonContainer: {
    display: 'flex',
    justifyContent: 'center',
  },
})

export default function CalenderSuccessForm(props) {
  const classes = useStyles()
  return (
    <Grid container spacing={2}>
      <Grid item xs={12} className={classes.subscribeButtonContainer}>
        <Button
          color='primary'
          variant='contained'
          href={'webcal://' + props.url}
        >
          Subscribe
        </Button>
      </Grid>
      <Grid item xs={12}>
        <Typography>
          <CopyText title={props.url} value={props.url} />
        </Typography>
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
