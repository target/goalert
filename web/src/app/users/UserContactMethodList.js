import React, { useState } from 'react'
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
import UserContactMethodVerificationDialog from './UserContactMethodVerificationDialog'

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

const DISABLED_CM_TOOLTIP =
  'Contact method disabled. See its options to reactivate.'

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
              action: this.props.readOnly ? null : (
                <Actions contactMethod={cm} />
              ),
              icon: cm.disabled ? (
                <Warning message={DISABLED_CM_TOOLTIP} data-cy='cm-disabled' />
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
      </Grid>
    )
  }
}

function Actions(props) {
  const { disabled, id } = props.contactMethod
  const [showEditDialog, setShowEditDialog] = useState(false)
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)
  const [showVerifyDialog, setShowVerifyDialog] = useState(false)

  function getTestOrVerifyAction(commit, disabled) {
    if (disabled) {
      return {
        label: 'Verify/Reactivate',
        onClick: () => setShowVerifyDialog(true),
      }
    } else {
      return {
        label: 'Send Test',
        onClick: () => commit(),
      }
    }
  }

  return (
    <Mutation mutation={testCM} client={graphql2Client} variables={{ id }}>
      {commit => (
        <React.Fragment>
          <OtherActions
            actions={[
              { label: 'Edit', onClick: () => setShowEditDialog(true) },
              {
                label: 'Delete',
                onClick: () => setShowDeleteDialog(true),
              },
              getTestOrVerifyAction(commit, disabled),
            ]}
          />
          {showEditDialog && (
            <UserContactMethodEditDialog
              contactMethodID={id}
              onClose={() => setShowEditDialog(false)}
            />
          )}
          {showDeleteDialog && (
            <UserContactMethodDeleteDialog
              contactMethodID={id}
              onClose={() => setShowDeleteDialog(false)}
            />
          )}
          {showVerifyDialog && (
            <UserContactMethodVerificationDialog
              contactMethodID={id}
              onClose={() => setShowVerifyDialog(false)}
            />
          )}
        </React.Fragment>
      )}
    </Mutation>
  )
}

Actions.propTypes = {
  contactMethod: p.shape({
    id: p.string.isRequired,
    disabled: p.bool.isRequired,
  }).isRequired,
}
