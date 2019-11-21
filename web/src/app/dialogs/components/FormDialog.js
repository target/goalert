import React, { Component } from 'react'
import { PropTypes as p } from 'prop-types'
import classnames from 'classnames'
import Button from '@material-ui/core/Button'
import Dialog from '@material-ui/core/Dialog'
import DialogActions from '@material-ui/core/DialogActions'
import DialogContent from '@material-ui/core/DialogContent'
import DialogContentText from '@material-ui/core/DialogContentText'
import Typography from '@material-ui/core/Typography'
import withStyles from '@material-ui/core/styles/withStyles'
import withMobileDialog from '@material-ui/core/withMobileDialog'
import LoadingButton from '../../loading/components/LoadingButton'
import { styles } from '../../styles/materialStyles'
import { DefaultTransition, FullscreenTransition } from '../../util/Transitions'
import DialogContentError from './DialogContentError'
import DialogTitleWrapper from './DialogTitleWrapper'

@withStyles(styles)
@withMobileDialog()
export default class FormDialog extends Component {
  static propTypes = {
    quickAction: p.shape({
      fields: p.object.isRequired,
      onRequestClose: p.func.isRequired,
      onSubmit: p.func.isRequired,
      onSuccess: p.func,
      title: p.string.isRequired,
      transition: p.func,
      subtitle: p.string,
      caption: p.string,
      contentOnly: p.bool,
      disableCancel: p.bool,
      errorMessage: p.string,
      readOnly: p.bool,
    }),
  }

  constructor(props) {
    super(props)
    this.state = {
      error: '',
      attemptCount: 0,
      loading: false,
    }
  }

  submitForm(e) {
    e.preventDefault()
    if (this.state.loading) return // dont allow multiple submissions while loading
    this.setState({ loading: true, error: '' })

    const result = this.props.onSubmit(e)

    if (!result || typeof result === 'string') {
      // if not a promise reject
      this.setState({
        error: result,
        attemptCount: this.state.attemptCount + 1,
        loading: false,
      })

      if (this.props.allowEdits) this.props.allowEdits()

      return
    }

    return result
      .then(args => {
        this.handleClose(true) // successful action
        if (this.props.onSuccess) this.props.onSuccess(args) // If the function exists run it
      })
      .catch(err => {
        this.setState({
          error: err.message || err,
          attemptCount: this.state.attemptCount + 1,
          loading: false,
        })

        if (this.props.allowEdits) this.props.allowEdits()
      })
  }

  handleClose = (successful = false, clickaway) => {
    if (this.state.loading && !successful) return
    this.props.onRequestClose(successful, clickaway)
  }

  render() {
    const loading = this.state.loading
    const {
      classes,
      title,
      fields,
      fullScreen,
      open,
      readOnly,
      subtitle,
      caption,
      contentOnly,
    } = this.props

    let titleJSX
    if (title) {
      titleJSX = (
        <DialogTitleWrapper
          key='title'
          fullScreen={fullScreen}
          title={title}
          onClose={this.handleClose}
        />
      )
    }

    let subtitleJSX
    if (subtitle) {
      subtitleJSX = (
        <DialogContent className={classes.defaultFlex} key='subtitle'>
          <DialogContentText>{subtitle}</DialogContentText>
        </DialogContent>
      )
    }

    let captionJSX
    if (caption) {
      captionJSX = (
        <DialogContent className={classes.defaultFlex} key='caption'>
          <Typography variant='caption'>{caption}</Typography>
        </DialogContent>
      )
    }

    const content = [
      titleJSX,
      subtitleJSX,
      <form
        key='form'
        onSubmit={e => this.submitForm(e)}
        style={{ width: '100%' }}
      >
        <DialogContent
          className={classnames(classes.defaultFlex, classes.overflowVisible)}
        >
          {fields}
        </DialogContent>
        {captionJSX}
        <DialogContentError
          className={classes.defaultFlex}
          error={this.props.errorMessage || this.state.error}
          noPadding
        />
        <DialogActions>
          <Button
            className={classes.cancelButton}
            onClick={this.handleClose}
            disabled={loading}
          >
            Cancel
          </Button>
          <LoadingButton
            attemptCount={this.state.attemptCount}
            buttonText='Submit'
            color='primary'
            disabled={readOnly}
            loading={loading}
            onClick={e => {
              this.submitForm(e)
            }}
          />
        </DialogActions>
      </form>,
    ]

    if (contentOnly && open) {
      return content
    }

    return (
      <Dialog
        open={open || false}
        onClose={() => this.handleClose(false, true)}
        classes={{
          paper: classnames(classes.dialogWidth, classes.overflowVisible),
        }}
        fullScreen={fullScreen}
        TransitionComponent={
          fullScreen ? FullscreenTransition : DefaultTransition
        }
        onExited={() => {
          if (this.props.resetForm) this.props.resetForm()
          this.setState({ error: '', attemptCount: 0, loading: false })
        }}
      >
        {content}
      </Dialog>
    )
  }
}
