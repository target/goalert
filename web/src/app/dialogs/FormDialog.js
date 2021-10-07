import React, { useState } from 'react'
import p from 'prop-types'
import Button from '@material-ui/core/Button'
import Dialog from '@material-ui/core/Dialog'
import DialogActions from '@material-ui/core/DialogActions'
import DialogContent from '@material-ui/core/DialogContent'
import Typography from '@material-ui/core/Typography'
import { makeStyles } from '@material-ui/core/styles'
import { isWidthUp } from '@material-ui/core/withWidth/index'

import { FadeTransition, SlideTransition } from '../util/Transitions'
import LoadingButton from '../loading/components/LoadingButton'
import DialogTitleWrapper from './components/DialogTitleWrapper'
import DialogContentError from './components/DialogContentError'
import { styles as globalStyles } from '../styles/materialStyles'
import { Form } from '../forms'
import ErrorBoundary from '../main/ErrorBoundary'
import Notices from '../details/Notices'
import useWidth from '../util/useWidth'

const useStyles = makeStyles((theme) => {
  const { cancelButton, dialogWidth } = globalStyles(theme)
  return {
    cancelButton,
    dialogWidth,
    form: {
      height: '100%', // pushes caption to bottom if room is available
    },
    dialogContent: {
      height: '100%', // parents of form need height set to properly function in Safari
      padding: 0,
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

function FormDialog(props) {
  const {
    alert,
    confirm,
    disableGutters,
    errors,
    fullScreen,
    loading,
    primaryActionLabel, // remove from dialogProps spread
    maxWidth,
    notices,
    onClose,
    onSubmit,
    subTitle, // can't be used in dialogProps spread
    title,
    onNext,
    onBack,
    fullHeight,
    ...dialogProps
  } = props

  const classes = useStyles()
  const width = useWidth()
  const isWideScreen = isWidthUp('md', width)
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

    let Component = DialogContent
    if (disableGutters) Component = 'div'

    return <Component className={classes.form}>{form}</Component>
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
    return props.errors.map((err, idx) => (
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
          <Button color='primary' onClick={handleOnClose} variant='contained'>
            {primaryActionLabel || 'Okay'}
          </Button>
        </DialogActions>
      )
    }

    const submitText = onNext ? 'Next' : 'Submit'

    return (
      <DialogActions>
        <Button
          className={classes.cancelButton}
          disabled={loading}
          onClick={onBack || handleOnClose}
        >
          {onBack ? 'Back' : 'Cancel'}
        </Button>
        <LoadingButton
          form='dialog-form'
          onClick={() => {
            if (!onNext) {
              setAttemptCount(attemptCount + 1)
            }
          }}
          attemptCount={attemptCount}
          buttonText={primaryActionLabel || (confirm ? 'Confirm' : submitText)}
          color='primary'
          loading={loading}
          type='submit'
        />
      </DialogActions>
    )
  }

  const fs = fullScreen || (!isWideScreen && !confirm)
  return (
    <Dialog
      classes={classesProp}
      fullScreen={fs}
      maxWidth={maxWidth}
      fullWidth
      open={open}
      onClose={(event, reason) => {
        if (reason === 'backdropClick' && (!isWideScreen || alert)) {
          // disable backdrop for mobile and alerts
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
      <DialogContent className={classes.dialogContent}>
        <Form
          id='dialog-form'
          className={classes.formContainer}
          onSubmit={(e, valid) => {
            e.preventDefault()
            if (valid) {
              onNext ? onNext() : onSubmit()
            }
          }}
        >
          <ErrorBoundary>{renderForm()}</ErrorBoundary>
        </Form>
      </DialogContent>
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

  errors: p.arrayOf(
    // this is an Error interface
    p.shape({
      message: p.string.isRequired,
    }),
  ),

  form: p.element,
  loading: p.bool,
  alert: p.bool,
  confirm: p.bool,
  maxWidth: p.string,

  // disables form content padding
  disableGutters: p.bool,

  // overrides any of the main action button titles with this specific text
  primaryActionLabel: p.string,

  // Callback fired when the dialog has exited.
  onClose: p.func,

  onSubmit: p.func,

  // if onNext is specified the submit button will be replaced with a 'Next' button
  onNext: p.func,
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
}

FormDialog.defaultProps = {
  errors: [],
  onClose: () => {},
  onSubmit: () => {},
  loading: false,
  confirm: false,
  caption: '',
  maxWidth: 'sm',
}

export default FormDialog
