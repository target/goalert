import React from 'react'
import { urlParamSelector } from '../selectors'
import { connect } from 'react-redux'
import { resetURLParams } from '../actions'
import { Dialog, DialogContent, DialogActions } from '@material-ui/core'
import DialogTitleWrapper from '../dialogs/components/DialogTitleWrapper'
import DialogContentError from '../dialogs/components/DialogContentError'
import LoadingButton from '../loading/components/LoadingButton'

@connect(
  (state) => ({
    errorMessage: urlParamSelector(state)('errorMessage'),
    errorTitle: urlParamSelector(state)('errorTitle'),
  }),
  (dispatch) => ({
    resetError: () => dispatch(resetURLParams('errorMessage', 'errorTitle')),
  }),
)
export default class URLErrorDialog extends React.Component {
  handleDialogClose = () => {
    this.props.resetError()
  }

  render() {
    const { errorMessage, errorTitle } = this.props
    const open = Boolean(errorMessage) || Boolean(errorTitle)

    return (
      open && (
        <Dialog
          open={open}
          onClose={(_, r) => {
              r !== 'backdropClick' ? this.onClose() : null
          }}
          aria-labelledby='alert-dialog-title'
          aria-describedby='alert-dialog-description'
        >
          <DialogTitleWrapper id='alert-dialog-title' title={errorTitle} />
          <DialogContent>
            <DialogContentError
              id='alert-dialog-description'
              error={errorMessage}
              noPadding
            />
          </DialogContent>
          <DialogActions>
            <LoadingButton
              buttonText='Okay'
              color='primary'
              onClick={() => this.onClose()}
            />
          </DialogActions>
        </Dialog>
      )
    )
  }
}
