import React, { Component } from 'react'
import { PropTypes as p } from 'prop-types'
import Button from '@material-ui/core/Button'
import Dialog from '@material-ui/core/Dialog'
import DialogActions from '@material-ui/core/DialogActions'
import DialogContent from '@material-ui/core/DialogContent'
import DialogTitle from '@material-ui/core/DialogTitle'
import Typography from '@material-ui/core/Typography'
import withStyles from '@material-ui/core/styles/withStyles'
import LoadingButton from '../../loading/components/LoadingButton'
import { styles } from '../../styles/materialStyles'
import DialogContentError from './DialogContentError'
import { Mutation } from 'react-apollo'

@withStyles(styles)
export default class ConfirmationDialog extends Component {
  static propTypes = {
    mutation: p.object,
    mutationVariables: p.object,
    onMutationSuccess: p.func,
    open: p.bool.isRequired,
    onRequestClose: p.func.isRequired,
    message: p.string,
    warning: p.string,
  }

  constructor(props) {
    super(props)
    this.state = {
      error: '',
      attemptCount: 0,
      loading: false,
    }
  }

  componentWillReceiveProps(nextProps) {
    if (!this.props.open && nextProps.open) {
      this.setState({ error: '', attemptCount: 0, loading: false })
    }
  }

  /*
   * Generic submit to handle actions with onClick functions.
   * Generally used for relay mutations
   */
  confirmAction() {
    if (typeof this.props.completeAction !== 'function') return

    this.props
      .completeAction()
      .then(() => {
        if (this._mnt) {
          // only set state if still mounted
          this.setState({
            error: '',
            attemptCount: 0,
            loading: false,
          })
        }
        this.props.onRequestClose(true) // successful action
      })
      .catch(err => this.handleError(err))
  }

  handleError = err => {
    this.setState({
      error: err.message || err,
      attemptCount: this.state.attemptCount + 1,
      loading: false,
    })
  }

  componentDidMount() {
    this._mnt = true
  }

  componentWillUnmount() {
    this._mnt = false
  }

  onMutationSubmit = (e, mutation) => {
    e.preventDefault()
    if (this.state.loading) return // dont allow multiple submissions while loading
    this.setState({ loading: true, error: null })
    return mutation({ variables: this.props.mutationVariables }).catch(error =>
      this.handleError(error.message),
    )
  }

  /*
   * Used for Apollo mutations as they are wrapped with a higher order component
   * to submit a mutation
   */
  renderAsMutation = () => {
    const { mutation, onMutationSuccess, onRequestClose } = this.props

    return (
      <Mutation
        key='mutation-form'
        mutation={mutation}
        refetchQueries={this.props.refetchQueries}
        update={(cache, { data }) => {
          this.setState({ loading: false })
          onRequestClose(true) // success = true prevents no-op set state in some funcs
          if (typeof onMutationSuccess === 'function') {
            onMutationSuccess(cache, data)
          }
        }}
      >
        {mutation => (
          <LoadingButton
            attemptCount={this.state.attemptCount}
            buttonText='Confirm'
            color='primary'
            loading={this.state.loading}
            onClick={e => this.onMutationSubmit(e, mutation)}
          />
        )}
      </Mutation>
    )
  }

  renderSubmit = () => {
    return (
      <LoadingButton
        attemptCount={this.state.attemptCount}
        buttonText='Confirm'
        color='primary'
        loading={this.state.loading}
        onClick={() => {
          this.setState({ loading: true, error: null })
          this.confirmAction()
        }}
      />
    )
  }

  render() {
    const {
      open,
      onRequestClose,
      classes,
      message,
      mutation,
      warning,
    } = this.props
    const { loading } = this.state

    return (
      <Dialog
        open={open}
        onClose={() => onRequestClose()}
        classes={{
          paper: classes.dialogWidth,
        }}
      >
        <DialogTitle>Are you sure?</DialogTitle>
        <DialogContent>
          <Typography style={{ width: '100%' }} variant='body1'>
            {message}
          </Typography>
          <Typography variant='caption'>
            <i>{warning}</i>
          </Typography>
        </DialogContent>
        <DialogContentError error={this.state.error} noPadding />
        <DialogActions>
          <Button
            className={this.props.classes.cancelButton}
            onClick={() => onRequestClose()}
            disabled={loading}
          >
            Cancel
          </Button>
          {mutation ? this.renderAsMutation() : this.renderSubmit()}
        </DialogActions>
      </Dialog>
    )
  }
}
