import React from 'react'
import p from 'prop-types'
import withStyles from '@material-ui/core/styles/withStyles'
import Button from '@material-ui/core/Button'
import Dialog from '@material-ui/core/Dialog'
import DialogActions from '@material-ui/core/DialogActions'
import DialogContent from '@material-ui/core/DialogContent'
import Typography from '@material-ui/core/Typography'
import withMobileDialog from '@material-ui/core/withMobileDialog'
import { DefaultTransition, FullscreenTransition } from '../util/Transitions'
import withWidth, { isWidthUp } from '@material-ui/core/withWidth/index'
import LoadingButton from '../loading/components/LoadingButton'
import DialogTitleWrapper from './components/DialogTitleWrapper'
import DialogContentError from './components/DialogContentError'
import { styles as globalStyles } from '../styles/materialStyles'
import gracefulUnmount from '../util/gracefulUnmount'
import { Form } from '../forms'

const styles = theme => {
  const { cancelButton, dialogWidth } = globalStyles(theme)
  return {
    cancelButton,
    dialogWidth,
    form: {
      height: '100%', // pushes caption to bottom if room is available
    },
    formContainer: {
      width: '100%',
      height: '100%',
      display: 'flex',
      flexDirection: 'column',
    },
    noGrow: {
      flexGrow: 0,
    },
  }
}

@withStyles(styles)
@withMobileDialog()
@withWidth()
@gracefulUnmount()
export default class FormDialog extends React.PureComponent {
  static propTypes = {
    title: p.string.isRequired,
    subTitle: p.string,
    caption: p.string,

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

    onClose: p.func,
    onSubmit: p.func,

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
      maxWidth,
      onClose,
      onSubmit,
      subTitle, // can't be used in dialogProps spread
      title,
      width,
      ...dialogProps
    } = this.props
    const isWideScreen = isWidthUp('md', width)

    return (
      <Dialog
        {...dialogProps}
        disableBackdropClick={!isWideScreen}
        fullScreen={!isWideScreen && !confirm && !alert}
        maxWidth={maxWidth}
        fullWidth
        open={!isUnmounting}
        onClose={alert ? null : onClose}
        TransitionComponent={
          isWideScreen || confirm || alert
            ? DefaultTransition
            : FullscreenTransition
        }
      >
        <DialogTitleWrapper
          fullScreen={!isWideScreen && !confirm && !alert}
          onClose={onClose}
          title={title}
          subTitle={subTitle}
        />
        <Form
          className={classes.formContainer}
          onSubmit={(e, valid) => {
            e.preventDefault()
            if (valid) onSubmit()
          }}
        >
          {this.renderForm()}
          {this.renderCaption()}
          {this.renderErrors()}
          {this.renderActions()}
        </Form>
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
        className={this.props.classes.noGrow}
        error={err.message || err}
        key={idx}
        noPadding
      />
    ))
  }

  renderActions = () => {
    const { alert, confirm, classes, errors, loading, onClose } = this.props

    if (alert) {
      return (
        <DialogActions>
          <Button color='primary' onClick={onClose} variant='contained'>
            Okay
          </Button>
        </DialogActions>
      )
    }

    return (
      <DialogActions>
        <Button
          className={classes.cancelButton}
          disabled={loading}
          onClick={onClose}
        >
          Cancel
        </Button>
        <LoadingButton
          attemptCount={errors.filter(e => !e.nonSubmit).length ? 1 : 0}
          buttonText={confirm ? 'Confirm' : 'Submit'}
          color='primary'
          loading={loading}
          type='submit'
        />
      </DialogActions>
    )
  }
}
