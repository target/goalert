import React from 'react'
import { PropTypes as p } from 'prop-types'
import {
  Button,
  FormHelperText,
  Grid,
  Typography,
  makeStyles,
} from '@material-ui/core'
import OpenInNewIcon from '@material-ui/icons/OpenInNew'
import CopyText from '../../util/CopyText'

const useStyles = makeStyles((theme) => ({
  caption: {
    width: '100%',
  },
  newTabIcon: {
    marginLeft: theme.spacing(1),
  },
  subscribeButtonContainer: {
    display: 'flex',
    justifyContent: 'center',
  },
}))

export default function CalenderSuccessForm(props) {
  const classes = useStyles()
  const url = props.url.replace(/^https?:\/\//, 'webcal://')
  return (
    <Grid container spacing={2}>
      <Grid item xs={12} className={classes.subscribeButtonContainer}>
        <Button
          color='primary'
          variant='contained'
          href={url}
          target='_blank'
          rel='noopener noreferrer'
        >
          Subscribe
          <OpenInNewIcon fontSize='small' className={classes.newTabIcon} />
        </Button>
      </Grid>
      <Grid item xs={12}>
        <Typography>
          <CopyText title={props.url} value={props.url} placement='bottom' />
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
