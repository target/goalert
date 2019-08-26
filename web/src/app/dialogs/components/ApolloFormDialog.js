import React, { Component } from 'react'
import { PropTypes as p } from 'prop-types'
import Button from '@material-ui/core/Button'
import Dialog from '@material-ui/core/Dialog'
import DialogActions from '@material-ui/core/DialogActions'
import DialogContent from '@material-ui/core/DialogContent'
import Typography from '@material-ui/core/Typography'
import withStyles from '@material-ui/core/styles/withStyles'
import withMobileDialog from '@material-ui/core/withMobileDialog'
import LoadingButton from '../../loading/components/LoadingButton'
import { styles } from '../../styles/materialStyles'
import { DefaultTransition, FullscreenTransition } from '../../util/Transitions'
import DialogContentError from './DialogContentError'
import DialogTitleWrapper from './DialogTitleWrapper'
import { Mutation } from 'react-apollo'
import classnames from 'classnames'
import { LegacyGraphQLClient } from '../../apollo'

/**
 * Consumes an apollo mutation (with an updater function, if applicable)
 * and renders a dialog with a form, as specified through props.
 *
 * The form submits to the mutation provided.
 */
@withStyles(styles)
@withMobileDialog()
export default class ApolloFormDialog extends Component {
  static propTypes = {
    quickAction: p.shape({
      allowEdits: p.func,
      fields: p.object.isRequired,
      getVariables: p.func,
      onRequestClose: p.func.isRequired,
      title: p.string.isRequired,
      transition: p.func,
      subtitle: p.string,
      caption: p.string,
      contentOnly: p.bool,
      disableCancel: p.bool,
      mutation: p.object.isRequired,
      shouldSubmit: p.func.isRequired,
      updater: p.func,
    }),
  }

  state = {
    error: '',
    attemptCount: 0,
    loading: false,
  }

  onError = error => {
    this.setState({
      attemptCount: this.state.attemptCount + 1,
      error,
      loading: false,
    })

    if (typeof this.props.allowEdits === 'function') this.props.allowEdits()
  }

  onSubmit = (e, mutation) => {
    e.preventDefault()
    if (this.state.loading) return // dont allow multiple submissions while loading
    const shouldSubmit = this.props.shouldSubmit() // validate fields, set to readOnly while committing, etc
    if (shouldSubmit) {
      this.setState({ loading: true })
      return mutation({ variables: this.props.getVariables() }).catch(error =>
        this.onError(error.message),
      )
    }
  }

  render() {
    const loading = this.state.loading
    const {
      caption,
      classes,
      contentOnly,
      fields,
      fullScreen,
      mutation,
      onRequestClose,
      onSuccess,
      open,
      subtitle,
      title,
    } = this.props

    let titleJSX
    if (title) {
      titleJSX = (
        <DialogTitleWrapper
          key='title'
          fullScreen={fullScreen}
          title={title}
          onClose={onRequestClose}
        />
      )
    }

    let subtitleJSX
    if (subtitle) {
      subtitleJSX = (
        <DialogContent key='subtitle' style={{ paddingBottom: 0 }}>
          <Typography variant='subtitle1'>{subtitle}</Typography>
        </DialogContent>
      )
    }

    let captionJSX
    if (caption) {
      captionJSX = (
        <DialogContent key='caption'>
          <Typography variant='caption'>{caption}</Typography>
        </DialogContent>
      )
    }

    const content = [
      titleJSX,
      subtitleJSX,
      <Mutation
        key='mutation-form'
        client={LegacyGraphQLClient}
        mutation={mutation}
        update={(cache, { data }) => {
          this.setState({ loading: false })
          onRequestClose()
          if (typeof onSuccess === 'function') {
            onSuccess(cache, data)
          }
        }}
      >
        {mutation => (
          <form
            onSubmit={e => this.onSubmit(e, mutation)}
            style={{ width: '100%' }}
          >
            <DialogContent className={classes.overflowVisible}>
              {fields}
            </DialogContent>
            {captionJSX}
            <DialogContentError error={this.state.error} noPadding />
            <DialogActions>
              <Button
                className={classes.cancelButton}
                onClick={onRequestClose}
                disabled={loading}
              >
                Cancel
              </Button>
              <LoadingButton
                attemptCount={this.state.attemptCount}
                buttonText='Submit'
                color='primary'
                type='submit'
                loading={loading}
              />
            </DialogActions>
          </form>
        )}
      </Mutation>,
    ]

    if (contentOnly && open) {
      return content
    }

    return (
      <Dialog
        open={open || false}
        onClose={onRequestClose}
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
