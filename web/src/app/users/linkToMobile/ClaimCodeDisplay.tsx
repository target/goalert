import React, { useEffect, useState } from 'react'
import {
  DialogContent,
  DialogContentText,
  Fade,
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
  claimCode: {
    letterSpacing: '2.5px',
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
  const [showManualCode, setShowManualCode] = useState(false)

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
          {!showManualCode && (
            <Typography
              className={classes.manualCodeTypography}
              variant='caption'
              color='textSecondary'
              onClick={() => setShowManualCode(true)}
            >
              Code not scanning? Click here to enter manually.
            </Typography>
          )}
          {showManualCode && (
            <Fade in={showManualCode}>
              <Typography
                className={classes.claimCode}
                variant='caption'
                color='textSecondary'
              >
                {props.claimCode}
              </Typography>
            </Fade>
          )}
        </Grid>
      </Grid>
    </DialogContent>
  )
}
