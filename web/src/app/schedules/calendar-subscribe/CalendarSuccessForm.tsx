import React from 'react'
import { Button, FormHelperText, Grid, Typography } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import OpenInNewIcon from '@mui/icons-material/OpenInNew'
import CopyText from '../../util/CopyText'
import { Theme } from '@mui/material/styles'

const useStyles = makeStyles((theme: Theme) => ({
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

interface CalendarSuccessFormProps {
  url: string
}

export default function CalenderSuccessForm({
  url,
}: CalendarSuccessFormProps): React.ReactNode {
  const classes = useStyles()
  const convertedUrl = url.replace(/^https?:\/\//, 'webcal://')
  return (
    <Grid container spacing={2}>
      <Grid item xs={12} className={classes.subscribeButtonContainer}>
        <Button
          variant='contained'
          href={convertedUrl}
          target='_blank'
          rel='noopener noreferrer'
        >
          Subscribe
          <OpenInNewIcon fontSize='small' className={classes.newTabIcon} />
        </Button>
      </Grid>
      <Grid item xs={12}>
        <Typography>
          <CopyText title={url} value={url} placement='bottom' asURL />
        </Typography>
        <FormHelperText>
          Some applications require you copy and paste the URL directly
        </FormHelperText>
      </Grid>
    </Grid>
  )
}
