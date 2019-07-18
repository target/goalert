import React from 'react'
import p from 'prop-types'
import gql from 'graphql-tag'
import Query from '../util/Query'
import DetailsPage from '../details/DetailsPage'
import StatusUpdateNotification from './UserStatusUpdatePreference'
import { UserAvatar } from '../util/avatar'
import UserContactMethodList from './UserContactMethodList'
import { AddAlarm, SettingsPhone } from '@material-ui/icons'
import SpeedDial from '../util/SpeedDial'
import UserNotificationRuleList from './UserNotificationRuleList'
import { Grid } from '@material-ui/core'
import UserContactMethodCreateDialog from './UserContactMethodCreateDialog'
import UserNotificationRuleCreateDialog from './UserNotificationRuleCreateDialog'
import Typography from '@material-ui/core/Typography'
import withStyles from '@material-ui/core/styles/withStyles'
import UserContactMethodVerificationDialog from './UserContactMethodVerificationDialog'

const styles = theme => ({
  profileImage: {
    width: 128,
    height: 128,
    margin: 'auto',
  },
})

const query = gql`
  query userInfo($id: ID!) {
    user(id: $id) {
      id
      name
      email
      onCallSteps {
        id
        escalationPolicy {
          id
          assignedTo {
            id
            name
          }
        }
      }
    }
  }
`

function serviceCount(onCallSteps = []) {
  const svcs = {}
  ;(onCallSteps || []).forEach(s =>
    (s.escalationPolicy.assignedTo || []).forEach(svc => (svcs[svc.id] = true)),
  )

  return Object.keys(svcs).length
}

@withStyles(styles)
export default class UserDetails extends React.PureComponent {
  static propTypes = {
    userID: p.string.isRequired,
    readOnly: p.bool,
  }

  state = {
    createCM: false,
    createNR: false,
    showVerifyDialogByID: null,
  }

  render() {
    return (
      <Query
        query={query}
        variables={{ id: this.props.userID }}
        render={({ data }) => this.renderData(data.user)}
      />
    )
  }

  renderData = user => {
    const svcCount = serviceCount(user.onCallSteps)
    return (
      <React.Fragment>
        {this.props.readOnly ? null : (
          <SpeedDial
            label='Add Items'
            actions={[
              {
                label: 'Add Contact Method',
                icon: <SettingsPhone />,
                onClick: () => this.setState({ createCM: true }),
              },
              {
                label: 'Add Notification Rule',
                icon: <AddAlarm />,
                onClick: () => this.setState({ createNR: true }),
              },
            ]}
          />
        )}
        {this.state.createCM && (
          <UserContactMethodCreateDialog
            userID={this.props.userID}
            onClose={result => {
              this.setState({
                createCM: false,
                showVerifyDialogByID:
                  result && result.contactMethodID
                    ? result.contactMethodID
                    : null,
              })
            }}
          />
        )}
        {this.state.showVerifyDialogByID && (
          <UserContactMethodVerificationDialog
            contactMethodID={this.state.showVerifyDialogByID}
            onClose={() => this.setState({ showVerifyDialogByID: null })}
          />
        )}
        {this.state.createNR && (
          <UserNotificationRuleCreateDialog
            userID={this.props.userID}
            onClose={() => this.setState({ createNR: false })}
          />
        )}
        <DetailsPage
          title={user.name + (svcCount ? ' (On-Call)' : '')}
          details={user.email}
          icon={
            <React.Fragment>
              <UserAvatar
                userID={this.props.userID}
                className={this.props.classes.profileImage}
              />
              <Typography
                variant='caption'
                style={{ textAlign: 'center', paddingTop: '0.5em' }}
              >
                Provided by{' '}
                <a href='https://gravatar.com' target='_blank'>
                  Gravatar
                </a>
              </Typography>
            </React.Fragment>
          }
          links={[
            {
              label: 'On-Call Assignments',
              url: 'on-call-assignments',
              subText: svcCount
                ? `On-call for ${svcCount} service${svcCount > 1 ? 's' : ''}`
                : 'Not currently on-call',
            },
          ]}
          titleFooter={
            this.props.readOnly ? null : (
              <StatusUpdateNotification userID={this.props.userID} />
            )
          }
          pageFooter={
            <Grid container spacing={2}>
              <UserContactMethodList
                userID={this.props.userID}
                readOnly={this.props.readOnly}
              />
              <UserNotificationRuleList
                userID={this.props.userID}
                readOnly={this.props.readOnly}
              />
            </Grid>
          }
        />
      </React.Fragment>
    )
  }
}
