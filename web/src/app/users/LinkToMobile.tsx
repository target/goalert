import React, { useEffect, useState, ReactNode } from 'react'
import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  Grid,
  Typography,
  makeStyles,
  isWidthDown,
} from '@material-ui/core'
import PhonelinkIcon from '@material-ui/icons/Phonelink'
import QRCode from 'qrcode.react'
import SwipeableViews from 'react-swipeable-views'
import { virtualize, bindKeyboard } from 'react-swipeable-views-utils'
import DialogTitleWrapper from '../dialogs/components/DialogTitleWrapper'
import useWidth from '../util/useWidth'
import { styles as globalStyles } from '../styles/materialStyles'

interface SlideParams {
  index: number
  key: number
}

const VirtualizeAnimatedViews = bindKeyboard(virtualize(SwipeableViews))

const useStyles = makeStyles((theme) => {
  const { cancelButton } = globalStyles(theme)

  return {
    cancelButton,
    centerItemContent: {
      display: 'flex',
      justifyContent: 'center',
    },
    manualCodeTypography: {
      '&:hover': {
        cursor: 'pointer',
        textDecoration: 'underline',
      },
    },
    qrContentText: {
      marginBottom: 0,
    },
  }
})

export default function LinkToMobile(): JSX.Element {
  const classes = useStyles()
  const width = useWidth()
  const fullscreen = isWidthDown('md', width)
  const [showDialog, setShowDialog] = useState(false)
  const [index, setIndex] = useState(0)

  // todo: add useEffects changing index as things are updated
  const [scanSuccessful, setScanSuccessful] = useState(false)
  useEffect(() => {
    if (scanSuccessful) {
      setIndex(1)
    }
  }, [scanSuccessful])

  function slideRenderer({ index, key }: SlideParams): ReactNode {
    switch (index) {
      case 0:
        return <ClaimCodeDisplay key={key} />
      case 1:
        return <VerifyCodeField key={key} />
      case 2:
        return <Success key={key} />
      case 3:
        return <Retry key={key} />
      default:
        return null
    }
  }

  return (
    <React.Fragment>
      <Button
        color='primary'
        variant='contained'
        startIcon={<PhonelinkIcon />}
        onClick={() => setShowDialog(true)}
      >
        Link to Mobile
      </Button>
      <Dialog
        open={true}
        fullScreen={fullscreen}
        fullWidth={true}
        maxWidth='xs'
        onClose={() => setShowDialog(false)}
      >
        <DialogTitleWrapper title='Link to Mobile' />
        <VirtualizeAnimatedViews
          index={index}
          onChangeIndex={(i: number) => setIndex(i)}
          slideRenderer={slideRenderer}
        />
        <DialogActions>
          <Button
            className={classes.cancelButton}
            onClick={() => setShowDialog(false)}
          >
            Cancel
          </Button>
          <Button
            className={classes.cancelButton}
            onClick={() => setScanSuccessful(true)}
          >
            ->
          </Button>
        </DialogActions>
      </Dialog>
    </React.Fragment>
  )
}

function ClaimCodeDisplay() {
  const classes = useStyles()

  return (
    <DialogContent>
      <Grid container spacing={1}>
        <Grid item xs={12}>
          <DialogContentText className={classes.qrContentText}>
            Scan this QR code from the GoAlert app on your mobile device to
            authenticate.
          </DialogContentText>
        </Grid>
        <Grid className={classes.centerItemContent} item xs={12}>
          <QRCode value='test' />
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

function VerifyCodeField() {
  return null
}

function Success() {
  return null
}

function Retry() {
  return null
}
