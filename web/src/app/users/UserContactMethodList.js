import React from 'react'
import p from 'prop-types'
import Query from '../util/Query'
import gql from 'graphql-tag'
import FlatList from '../lists/FlatList'
import { Grid, Card, CardHeader } from '@material-ui/core'
import { formatCMValue, sortContactMethods } from './util'
import OtherActions from '../util/OtherActions'
import { Mutation } from 'react-apollo'
import { graphql2Client } from '../apollo'
import UserContactMethodDeleteDialog from './UserContactMethodDeleteDialog'
import UserContactMethodEditDialog from './UserContactMethodEditDialog'
import ListItem from '@material-ui/core/ListItem'
import ListItemText from '@material-ui/core/ListItemText'
import { Config } from '../util/RequireConfig'
import { Warning } from '../icons'
import ContactMethodVerificationDialog from './ContactMethodVerificationDialog'

const query = gql`
  query cmList($id: ID!) {
    user(id: $id) {
      id
      contactMethods {
        id
        name
        type
        value
        disabled
      }
    }
  }
`

const testCM = gql`
  mutation($id: ID!) {
    testContactMethod(id: $id)
  }
`

const DISABLED_CM_TOOLTIP = `Number disabled. See contact method's options to reactivate`

export default class UserContactMethodList extends React.PureComponent {
  static propTypes = {
    userID: p.string.isRequired,
    readOnly: p.bool,
  }

  state = {
    edit: null,
    delete: null,
    isVerifyDialogOpen: false,
  }

  render() {
    return (
      <Query
        query={query}
        variables={{ id: this.props.userID }}
        render={({ data }) => this.renderList(data.user.contactMethods)}
      />
    )
  }

  getTestOrVerifyAction = (commit, disabled) => {
    if (disabled) {
      return {
        label: 'Verify/Reactivate',
        onClick: () => this.setState({ isVerifyDialogOpen: true }),
      }
    }

    return {
      label: 'Send Test',
      onClick: () => commit(),
    }
  }

  renderActions(cm) {
    const { disabled, id } = cm

    return (
      <Mutation mutation={testCM} client={graphql2Client} variables={{ id }}>
        {commit => (
          <React.Fragment>
            <OtherActions
              actions={[
                { label: 'Edit', onClick: () => this.setState({ edit: id }) },
                {
                  label: 'Delete',
                  onClick: () => this.setState({ delete: id }),
                },
                this.getTestOrVerifyAction(commit, disabled),
              ]}
            />
            {this.state.isVerifyDialogOpen && (
              <ContactMethodVerificationDialog
                onClose={() => this.setState({ isVerifyDialogOpen: false })}
                contactMethodID={id}
              />
            )}
          </React.Fragment>
        )}
      </Mutation>
    )
  }

  renderList(contactMethods) {
    return (
      <Grid item xs={12}>
        <Card>
          <CardHeader title='Contact Methods' />
          <FlatList
            data-cy='contact-methods'
            items={sortContactMethods(contactMethods).map(cm => ({
              title: `${cm.name} (${cm.type})`,
              subText: formatCMValue(cm.type, cm.value),
              action: this.props.readOnly ? null : this.renderActions(cm),
              icon: cm.disabled ? (
                <Warning tooltip={DISABLED_CM_TOOLTIP} />
              ) : null,
            }))}
            emptyMessage='No contact methods'
          />
          <Config>
            {cfg =>
              !this.props.readOnly &&
              cfg['General.NotificationDisclaimer'] && (
                <ListItem>
                  <ListItemText
                    secondary={cfg['General.NotificationDisclaimer']}
                  />
                </ListItem>
              )
            }
          </Config>
        </Card>
        {this.state.edit && (
          <UserContactMethodEditDialog
            cmID={this.state.edit}
            onClose={() => this.setState({ edit: null })}
          />
        )}
        {this.state.delete && (
          <UserContactMethodDeleteDialog
            cmID={this.state.delete}
            onClose={() => this.setState({ delete: null })}
          />
        )}
      </Grid>
    )
  }
}
