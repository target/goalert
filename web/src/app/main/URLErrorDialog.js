import React from 'react'
import { urlParamSelector } from '../selectors'
import { connect } from 'react-redux'
import FormDialog from '../dialogs/FormDialog'
import { resetURLParams } from '../actions'

@connect(
  state => ({
    errorMessage: urlParamSelector(state)('errorMessage'),
    errorTitle: urlParamSelector(state)('errorTitle'),
  }),
  dispatch => ({
    resetError: () => dispatch(resetURLParams('errorMessage', 'errorTitle')),
  }),
)
export default class URLErrorDialog extends React.Component {
  onClose = () => {
    this.props.resetError()
  }

  render() {
    const { errorMessage, errorTitle } = this.props
    const open = Boolean(errorMessage) || Boolean(errorTitle)

    return (
      open && (
        <FormDialog
          alert
          errors={[
            {
              message: errorMessage || 'Oops! Something went wrong.',
            },
          ]}
          maxWidth='sm'
          onClose={this.onClose}
          title={errorTitle || 'An error occurred'}
        />
      )
    )
  }
}
