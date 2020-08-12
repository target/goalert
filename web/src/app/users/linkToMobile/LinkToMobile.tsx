import React, { useEffect, useState, ReactNode } from 'react'
import {
  Button,
  Dialog,
  DialogActions,
  makeStyles,
  isWidthDown,
} from '@material-ui/core'
import PhonelinkIcon from '@material-ui/icons/Phonelink'
import SwipeableViews from 'react-swipeable-views'
import { virtualize, bindKeyboard } from 'react-swipeable-views-utils'
import DialogTitleWrapper from '../../dialogs/components/DialogTitleWrapper'
import useWidth from '../../util/useWidth'
import { styles as globalStyles } from '../../styles/materialStyles'
import ClaimCodeDisplay from './ClaimCodeDisplay'
import VerifyCodeFields from './VerifyCodeFields'
// import Success from './Success'
// import VerifyCodeFields from './VerifyCodeFields'

interface SlideParams {
  index: number
  key: number
}

const VirtualizeAnimatedViews = bindKeyboard(virtualize(SwipeableViews))

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
  const [index, setIndex] = useState(1)

  // todo: add useEffects changing index as things are updated

  function slideRenderer({ index, key }: SlideParams): ReactNode {
    switch (index) {
      case 0:
        return <ClaimCodeDisplay key={key} />
      case 1:
        return <VerifyCodeFields key={key} />
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
        </DialogActions>
      </Dialog>
    </React.Fragment>
  )
}

function Success() {
  return null
}

function Retry() {
  return null
}
