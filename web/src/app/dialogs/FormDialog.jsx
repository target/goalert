import React, { useState } from 'react'
import p from 'prop-types'
import Button from '@mui/material/Button'
import Dialog from '@mui/material/Dialog'
import DialogActions from '@mui/material/DialogActions'
import DialogContent from '@mui/material/DialogContent'
import Tooltip from '@mui/material/Tooltip'
import Typography from '@mui/material/Typography'
import makeStyles from '@mui/styles/makeStyles'

import { FadeTransition, SlideTransition } from '../util/Transitions'
import LoadingButton from '../loading/components/LoadingButton'
import DialogTitleWrapper from './components/DialogTitleWrapper'
import DialogContentError from './components/DialogContentError'
import { styles as globalStyles } from '../styles/materialStyles'
import { Form } from '../forms'
import ErrorBoundary from '../main/ErrorBoundary'
import Notices from '../details/Notices'
import { useIsWidthUp } from '../util/useWidth'

const useStyles = makeStyles((theme) => {
  const { dialogWidth } = globalStyles(theme)
  return {
    dialogWidth,
    form: {
      height: '100%', // pushes caption to bottom if room is available
    },
    dialogContent: {
      height: '100%', // parents of form need height set to properly function in Safari
      paddingTop: '8px',
    },
    formContainer: {
      width: '100%',
      height: '100%',
      display: 'flex',
      flexDirection: 'column',
    },
    errorContainer: {
      flexGrow: 0,
      overflowY: 'visible',
    },
    fullHeight: {
      height: '100%',
    },
  }
})

export default function FormDialog(props) {
  const {
    alert,
    confirm = false,
    errors = [],
    fullScreen,
    loading = false,
    primaryActionLabel, // remove from dialogProps spread
    maxWidth = 'sm',
    notices,
    onClose = () => {},
    onSubmit = () => {},
    subTitle, // can't be used in dialogProps spread
    title,
    onNext,
    onBack,
    fullHeight,
    nextTooltip,
    disableBackdropClose,
    disablePortal,
    disableSubmit,
    disableNext,
    ...dialogProps
  } = props

  const classes = useStyles()
  const isWideScreen = useIsWidthUp('md')
  const [open, setOpen] = useState(true)
  const [attemptCount, setAttemptCount] = useState(0)

  const classesProp = fullHeight
    ? {
        paper: classes.fullHeight,
      }
    : {}

  const handleOnClose = () => {
    setOpen(false)
  }

  const handleOnExited = () => {
    onClose()
  }

  function renderForm() {
    const { form } = props

    // don't render empty space
    if (!form) {
      return null
    }

    return (
      <DialogContent className={classes.dialogContent}>
        <Form
          id='dialog-form'
          className={classes.formContainer}
          onSubmit={(e, valid) => {
            e.preventDefault()
            if (valid) {
              onSubmit()
            }
          }}
        >
          <ErrorBoundary>
            <div className={classes.form}>{form}</div>
          </ErrorBoundary>
        </Form>
      </DialogContent>
    )
  }

  function renderCaption() {
    if (!props.caption) return null

    return (
      <DialogContent>
        <Typography variant='caption'>{props.caption}</Typography>
      </DialogContent>
    )
  }

  function renderErrors() {
    const errors =
      typeof props.errors === 'function' ? props.errors() : props.errors
    return errors.map((err, idx) => (
      <DialogContentError
        className={classes.errorContainer}
        error={err.message || err}
        key={idx}
        noPadding
      />
    ))
  }

  function renderActions() {
    if (alert) {
      return (
        <DialogActions>
          <Button onClick={handleOnClose} variant='contained'>
            {primaryActionLabel || 'Okay'}
          </Button>
        </DialogActions>
      )
    }

    const nextButton = (
      <Button
        variant='contained'
        color='secondary'
        disabled={loading || disableNext}
        onClick={onNext}
        sx={{ mr: 1 }}
      >
        Next
      </Button>
    )

    return (
      <DialogActions>
        <Button
          disabled={loading}
          color='secondary'
          onClick={onBack || handleOnClose}
          sx={{ mr: 1 }}
        >
          {onBack ? 'Back' : 'Cancel'}
        </Button>

        {onNext && nextTooltip ? (
          <Tooltip title={nextTooltip}>
            {/* wrapping in span as button may be disabled */}
            <span>{nextButton}</span>
          </Tooltip>
        ) : onNext ? (
          nextButton
        ) : null}

        <LoadingButton
          form='dialog-form'
          onClick={() => {
            setAttemptCount(attemptCount + 1)

            if (!props.form) {
              onSubmit()
            }
          }}
          attemptCount={attemptCount}
          buttonText={primaryActionLabel || (confirm ? 'Confirm' : 'Submit')}
          disabled={loading || disableSubmit}
          loading={loading}
          type='submit'
        />
      </DialogActions>
    )
  }

  const fs = fullScreen || (!isWideScreen && !confirm)
  return (
    <Dialog
      disablePortal={disablePortal}
      classes={classesProp}
      fullScreen={fs}
      maxWidth={maxWidth}
      fullWidth
      open={open}
      onClose={(_, reason) => {
        if (
          reason === 'backdropClick' &&
          (!isWideScreen || alert || disableBackdropClose)
        ) {
          return
        }
        handleOnClose()
      }}
      TransitionComponent={
        isWideScreen || confirm ? FadeTransition : SlideTransition
      }
      {...dialogProps}
      TransitionProps={{
        onExited: handleOnExited,
      }}
    >
      <Notices notices={notices} />
      <DialogTitleWrapper
        fullScreen={fs}
        onClose={handleOnClose}
        title={title}
        subTitle={subTitle}
      />
      {renderForm()}
      {renderCaption()}
      {renderErrors()}
      {renderActions()}
    </Dialog>
  )
}

FormDialog.propTypes = {
  title: p.node.isRequired,
  subTitle: p.node,
  caption: p.node,

  errors: p.oneOfType([
    p.arrayOf(
      // this is an Error interface
      p.shape({
        message: p.string.isRequired,
      }),
    ),
    p.func,
  ]),

  form: p.node,
  loading: p.bool,
  alert: p.bool,
  confirm: p.bool,
  maxWidth: p.string,

  disablePortal: p.bool, // disable the portal behavior of the dialog

  disableNext: p.bool, // disables the next button while true
  disableSubmit: p.bool, // disables the submit button while true

  // overrides any of the main action button titles with this specific text
  primaryActionLabel: p.string,

  // Callback fired when the dialog has exited.
  onClose: p.func,

  onSubmit: p.func,

  // if onNext is specified the submit button will be replaced with a 'Next' button
  onNext: p.func,
  nextTooltip: p.string,

  // if onBack is specified the cancel button will be replaced with a 'Back' button
  onBack: p.func,

  // allow the dialog to grow beyond the normal max-width.
  grow: p.bool,

  // If true, the dialog will be full-screen
  fullScreen: p.bool,

  // notices to render; see details/Notices.tsx
  notices: p.arrayOf(p.object),

  // make dialog fill vertical space
  fullHeight: p.bool,

  disableBackdropClose: p.bool,

  // gets spread to material-ui
  PaperProps: p.object,
}
