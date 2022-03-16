import React from 'react'
import { PropTypes as p } from 'prop-types'
import { Button, FormHelperText, Grid, Typography } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import OpenInNewIcon from '@mui/icons-material/OpenInNew'
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
          <CopyText
            title={props.url}
            value={props.url}
            placement='bottom'
            asURL
          />
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
