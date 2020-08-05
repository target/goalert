import React, { useState } from 'react'
import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  Grid,
  Stepper,
  Step,
  makeStyles,
  isWidthDown,
} from '@material-ui/core'
import PhonelinkIcon from '@material-ui/icons/Phonelink'
import QRCode from 'qrcode.react'
import DialogTitleWrapper from '../dialogs/components/DialogTitleWrapper'
import useWidth from '../util/useWidth'
import { styles as globalStyles } from '../styles/materialStyles'

const useStyles = makeStyles((theme) => {
  const { cancelButton } = globalStyles(theme)

  return {
    cancelButton,
  }
})

export default function LinkToMobile(): JSX.Element {
  const classes = useStyles()
  const width = useWidth()
  const fullscreen = isWidthDown('md', width)
  const [showDialog, setShowDialog] = useState(false)

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
        open={showDialog}
        fullScreen={fullscreen}
        onClose={() => setShowDialog(false)}
      >
        <DialogTitleWrapper title='Link to Mobile' />
        <DialogContent>Test</DialogContent>
        <DialogActions>
          <Button
            className={classes.cancelButton}
            onClick={() => setShowDialog(false)}
          >
            Cancel
          </Button>
        </DialogActions>
      </Dialog>
    </React.Fragment>
  )
}
