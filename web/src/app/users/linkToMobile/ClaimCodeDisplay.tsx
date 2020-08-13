import React, { useEffect } from 'react'
import {
  DialogContent,
  DialogContentText,
  Grid,
  Typography,
  makeStyles,
} from '@material-ui/core'
import QRCode from 'qrcode.react'

const useStyles = makeStyles({
  centerItemContent: {
    display: 'flex',
    justifyContent: 'center',
  },
  contentText: {
    marginBottom: 0,
  },
  manualCodeTypography: {
    '&:hover': {
      cursor: 'pointer',
      textDecoration: 'underline',
    },
  },
})

interface ClaimCodeDisplayProps {
  authLinkID: string
  claimCode: string
}

export default function ClaimCodeDisplay(
  props: ClaimCodeDisplayProps,
): JSX.Element {
  const classes = useStyles()

  useEffect(() => {
    console.log('id: ', props.authLinkID)
    console.log('code: ', props.claimCode)
  }, [])

  return (
    <DialogContent>
      <Grid container spacing={1}>
        <Grid item xs={12}>
          <DialogContentText className={classes.contentText}>
            Scan this QR code from the GoAlert app on your mobile device to
            authenticate.
          </DialogContentText>
        </Grid>
        <Grid className={classes.centerItemContent} item xs={12}>
          <QRCode value={props.claimCode} />
        </Grid>
        <Grid className={classes.centerItemContent} item xs={12}>
          <Typography
            className={classes.manualCodeTypography}
            variant='caption'
            color='textSecondary'
          >
            Code not scanning? Click here to enter manually.
          </Typography>
        </Grid>
      </Grid>
    </DialogContent>
  )
}
