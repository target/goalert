import React, { useEffect, useState, ReactNode, ChangeEvent } from 'react'
import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  Grid,
  TextField,
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
    codeContainer: {
      marginTop: '2em',
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
    textField: {
      textAlign: 'center',
      fontSize: '1.75rem',
    },
  }
})

export default function LinkToMobile(): JSX.Element {
  const classes = useStyles()
  const width = useWidth()
  const fullscreen = isWidthDown('md', width)
  const [showDialog, setShowDialog] = useState(false)
  const [index, setIndex] = useState(1)

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
            {'->'}
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
          <DialogContentText className={classes.contentText}>
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
  const classes = useStyles()
  const [numOne, setNumOne] = useState('')
  const [numTwo, setNumTwo] = useState('')
  const [numThree, setNumThree] = useState('')
  const [numFour, setNumFour] = useState('')

  function handleChange(
    e: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
    curVal: string,
    setter: Function,
  ): void {
    setter(e.target.value)
  }

  return (
    <DialogContent>
      <Grid container spacing={2}>
        <Grid item xs={12}>
          <DialogContentText className={classes.contentText}>
            Enter the code displayed on your mobile device.
          </DialogContentText>
        </Grid>
        <Grid
          className={classes.codeContainer}
          item
          xs={12}
          container
          spacing={2}
        >
          <Grid item xs={3}>
            <TextField
              value={numOne}
              onChange={(e) => handleChange(e, numOne, setNumOne)}
              inputProps={{
                maxLength: 1,
                className: classes.textField,
              }}
            />
          </Grid>
          <Grid item xs={3}>
            <TextField
              value={numTwo}
              onChange={(e) => handleChange(e, numTwo, setNumTwo)}
              inputProps={{
                maxLength: 1,
                className: classes.textField,
              }}
            />
          </Grid>
          <Grid item xs={3}>
            <TextField
              value={numThree}
              onChange={(e) => handleChange(e, numThree, setNumThree)}
              inputProps={{
                maxLength: 1,
                className: classes.textField,
              }}
            />
          </Grid>
          <Grid item xs={3}>
            <TextField
              value={numFour}
              onChange={(e) => handleChange(e, numFour, setNumFour)}
              inputProps={{
                maxLength: 1,
                className: classes.textField,
              }}
            />
          </Grid>
        </Grid>
      </Grid>
    </DialogContent>
  )
}

function Success() {
  return null
}

function Retry() {
  return null
}
