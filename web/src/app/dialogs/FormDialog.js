import React from 'react'
import p from 'prop-types'
import Button from '@material-ui/core/Button'
import Dialog from '@material-ui/core/Dialog'
import DialogActions from '@material-ui/core/DialogActions'
import DialogContent from '@material-ui/core/DialogContent'
import Typography from '@material-ui/core/Typography'
import { makeStyles } from '@material-ui/core/styles'
import { isWidthUp } from '@material-ui/core/withWidth/index'

import { DefaultTransition, FullscreenTransition } from '../util/Transitions'
import LoadingButton from '../loading/components/LoadingButton'
import DialogTitleWrapper from './components/DialogTitleWrapper'
import DialogContentError from './components/DialogContentError'
import { styles as globalStyles } from '../styles/materialStyles'
import withGracefulUnmount from '../util/gracefulUnmount'
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
  }
})

function FormDialog(props) {
  const classes = useStyles()
  const width = useWidth()
  const isWideScreen = isWidthUp('md', width)
  const {
    alert,
    confirm,
    disableGutters,
    errors,
    fullScreen,
    isUnmounting,
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
    ...dialogProps
  } = props

  function renderForm() {
    const { disableGutters, form } = props

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
    const {
      alert,
      confirm,
      errors,
      loading,
      primaryActionLabel,
      onClose,
      onBack,
      onNext,
    } = props

    if (alert) {
      return (
        <DialogActions>
          <Button color='primary' onClick={onClose} variant='contained'>
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
          onClick={onBack || onClose}
        >
          {onBack ? 'Back' : 'Cancel'}
        </Button>
        <LoadingButton
          form='dialog-form'
          attemptCount={errors.filter((e) => !e.nonSubmit).length ? 1 : 0}
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
      disableBackdropClick={!isWideScreen || alert}
      fullScreen={fs}
      maxWidth={maxWidth}
      fullWidth
      open={!isUnmounting}
      onClose={onClose}
      TransitionComponent={fs ? FullscreenTransition : DefaultTransition}
      {...dialogProps}
    >
      <Notices notices={notices} />
      <DialogTitleWrapper
        fullScreen={fs}
        onClose={onClose}
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
    p.shape({
      message: p.string.isRequired,
      nonSubmit: p.bool, // indicates that it is a non-submit related error
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

  onClose: p.func,
  onSubmit: p.func,

  // if onNext is specified the submit button will be replaced with a 'Next' button
  onNext: p.func,
  // if onBack is specified the cancel button will be replaced with a 'Back' button
  onBack: p.func,

  // provided by graceful unmount
  isUnmounting: p.bool,
  onExited: p.func,

  // allow the dialog to grow beyond the normal max-width.
  grow: p.bool,
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

export default withGracefulUnmount(FormDialog)
