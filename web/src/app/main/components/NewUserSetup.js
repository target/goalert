import React, { Component } from 'react'
import p from 'prop-types'
import UserContactMethodCreateDialog from '../../users/UserContactMethodCreateDialog'
import UserContactMethodVerificationDialog from '../../users/UserContactMethodVerificationDialog'
import { clearParameter } from '../../util/query_param'
import { graphql } from 'react-apollo'
import gql from 'graphql-tag'
import { connect } from 'react-redux'
import { bindActionCreators } from 'redux'
import { withRouter } from 'react-router-dom'
import { setShowNewUserForm } from '../../actions'

const ID_QUERY = gql`
  query GetCurrentUserID {
    currentUser {
      id
      contact_methods {
        id
      }
      notification_rules {
        id
      }
    }
  }
`

const mapStateToProps = state => ({
  isFirstLogin: state.main.isFirstLogin,
})

const mapDispatchToProps = dispatch =>
  bindActionCreators(
    {
      setShowNewUserForm,
    },
    dispatch,
  )

@graphql(ID_QUERY)
@connect(
  mapStateToProps,
  mapDispatchToProps,
)
@withRouter
export default class NewUserSetup extends Component {
  static contextTypes = {
    isFirstLogin: p.bool,
    setShowNewUserForm: p.func,
  }

  state = {
    createCM: false,
    showVerifyDialogByID: null,
  }

  onExit(result) {
    this.setState({
      createCM: false,
      showVerifyDialogByID:
        result && result.contactMethodID ? result.contactMethodID : null,
    })

    this.onNewUserDialogClose()
  }

  /*
   * Don't show the new user setup dialog if the user keeps refreshing with
   * the original query param still active
   */
  onNewUserDialogClose = (successful, clickaway) => {
    const newUrl = clearParameter('isFirstLogin')
    this.props.history.replace(window.location.pathname + newUrl)
    if (clickaway) return
    this.props.setShowNewUserForm()
  }

  render() {
    const { data, isFirstLogin } = this.props
    const userID = data && data.currentUser && data.currentUser.id
    const title = 'Welcome to GoAlert!'
    const subtitle = 'To get started, please enter a contact method.'

    if (!userID) {
      return null
    }

    if (isFirstLogin) {
      return (
        <React.Fragment>
          <UserContactMethodCreateDialog
            title={title}
            subtitle={subtitle}
            userID={userID}
            onClose={result => {
              this.onExit(result)
            }}
          />
          {this.state.showVerifyDialogByID && (
            <UserContactMethodVerificationDialog
              contactMethodID={this.state.showVerifyDialogByID}
              onClose={() => this.setState({ showVerifyDialogByID: null })}
            />
          )}
        </React.Fragment>
      )
    }
  }
}
