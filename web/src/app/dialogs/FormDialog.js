import React from 'react'
import p from 'prop-types'
import withStyles from '@material-ui/core/styles/withStyles'
import Button from '@material-ui/core/Button'
import Dialog from '@material-ui/core/Dialog'
import DialogActions from '@material-ui/core/DialogActions'
import DialogContent from '@material-ui/core/DialogContent'
import Typography from '@material-ui/core/Typography'
import { DefaultTransition, FullscreenTransition } from '../util/Transitions'
import withWidth, { isWidthUp } from '@material-ui/core/withWidth/index'
import LoadingButton from '../loading/components/LoadingButton'
import DialogTitleWrapper from './components/DialogTitleWrapper'
import DialogContentError from './components/DialogContentError'
import { styles as globalStyles } from '../styles/materialStyles'
import gracefulUnmount from '../util/gracefulUnmount'
import { Form } from '../forms'
import ErrorBoundary from '../main/ErrorBoundary'

const styles = theme => {
  const { cancelButton, dialogWidth } = globalStyles(theme)
  return {
    cancelButton,
    dialogWidth,
    form: {
      height: '100%', // pushes caption to bottom if room is available
    },
    dialogContent: {
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
}

@withStyles(styles)
@withWidth()
@gracefulUnmount()
export default class FormDialog extends React.PureComponent {
  static propTypes = {
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

    // provided by gracefulUnmount()
    isUnmounting: p.bool,
    onExited: p.func,

    // allow the dialog to grow beyond the normal max-width.
    grow: p.bool,
  }

  static defaultProps = {
    errors: [],
    onClose: () => {},
    onSubmit: () => {},
    loading: false,
    confirm: false,
    caption: '',
    maxWidth: 'sm',
  }

  render() {
    const {
      alert,
      classes,
      confirm,
      disableGutters,
      errors,
      isUnmounting,
      loading,
      primaryActionLabel, // remove from dialogProps spread
      maxWidth,
      onClose,
      onSubmit,
      subTitle, // can't be used in dialogProps spread
      title,
      width,
      onNext,
      onBack,
      ...dialogProps
    } = this.props

    const isWideScreen = isWidthUp('md', width)

    return (
      <Dialog
        disableBackdropClick={!isWideScreen || alert}
        fullScreen={!isWideScreen && !confirm}
        maxWidth={maxWidth}
        fullWidth
        open={!isUnmounting}
        onClose={onClose}
        TransitionComponent={
          isWideScreen || confirm ? DefaultTransition : FullscreenTransition
        }
        {...dialogProps}
      >
        <DialogTitleWrapper
          fullScreen={!isWideScreen && !confirm}
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
            <ErrorBoundary>{this.renderForm()}</ErrorBoundary>
          </Form>
        </DialogContent>
        {this.renderCaption()}
        {this.renderErrors()}
        {this.renderActions()}
      </Dialog>
    )
  }

  renderForm = () => {
    const { classes, disableGutters, form } = this.props

    // don't render empty space
    if (!form) {
      return null
    }

    let Component = DialogContent
    if (disableGutters) Component = 'div'

    return <Component className={classes.form}>{form}</Component>
  }

  renderCaption = () => {
    if (!this.props.caption) return null

    return (
      <DialogContent>
        <Typography variant='caption'>{this.props.caption}</Typography>
      </DialogContent>
    )
  }

  renderErrors = () => {
    return this.props.errors.map((err, idx) => (
      <DialogContentError
        className={this.props.classes.errorContainer}
        error={err.message || err}
        key={idx}
        noPadding
      />
    ))
  }

  renderActions = () => {
    const {
      alert,
      confirm,
      classes,
      errors,
      loading,
      primaryActionLabel,
      onClose,
      onBack,
      onNext,
    } = this.props

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
          attemptCount={errors.filter(e => !e.nonSubmit).length ? 1 : 0}
          buttonText={primaryActionLabel || (confirm ? 'Confirm' : submitText)}
          color='primary'
          loading={loading}
          type='submit'
        />
      </DialogActions>
    )
  }
}
