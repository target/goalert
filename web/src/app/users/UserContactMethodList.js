import React, { useState } from 'react'
import p from 'prop-types'
import Query from '../util/Query'
import gql from 'graphql-tag'
import FlatList from '../lists/FlatList'
import { Button, Card, CardHeader, Grid, IconButton } from '@material-ui/core'
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
import withStyles from '@material-ui/core/styles/withStyles'

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

const styles = {
  actionGrid: {
    display: 'flex',
    alignItems: 'center',
  },
}

@withStyles(styles)
export default class UserContactMethodList extends React.PureComponent {
  static propTypes = {
    userID: p.string.isRequired,
    readOnly: p.bool,
  }

  state = {
    showVerifyDialogByID: null,
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
    const { readOnly } = this.props

    const getIcon = cm => {
      if (cm.disabled && readOnly) {
        return <Warning title='Contact method disabled' />
      } else if (cm.disabled && !readOnly) {
        return (
          <IconButton
            aria-label='Reactivate contact method'
            onClick={() => this.setState({ showVerifyDialogByID: cm.id })}
            variant='contained'
            color='primary'
            disabled={readOnly}
          >
            <Warning title='Contact method disabled' />
          </IconButton>
        )
      }
    }

    const getSecondaryAction = cm => {
      return (
        <Grid container spacing={2} className={this.props.classes.actionGrid}>
          {cm.disabled && !readOnly && (
            <Grid item>
              <Button
                aria-label='Reactivate contact method'
                onClick={() => this.setState({ showVerifyDialogByID: cm.id })}
                variant='contained'
                color='primary'
              >
                Reactivate
              </Button>
            </Grid>
          )}
          {!readOnly && (
            <Grid item>
              <Actions contactMethod={cm} />
            </Grid>
          )}
        </Grid>
      )
    }

    return (
      <Grid item xs={12}>
        <Card>
          <CardHeader title='Contact Methods' />
          <FlatList
            data-cy='contact-methods'
            items={sortContactMethods(contactMethods).map(cm => ({
              title: `${cm.name} (${cm.type})${
                cm.disabled ? ' - Disabled' : ''
              }`,
              subText: formatCMValue(cm.type, cm.value),
              secondaryAction: getSecondaryAction(cm),
              icon: getIcon(cm),
            }))}
            emptyMessage='No contact methods'
          />
          {this.state.showVerifyDialogByID && (
            <UserContactMethodVerificationDialog
              contactMethodID={this.state.showVerifyDialogByID}
              onClose={() => this.setState({ showVerifyDialogByID: null })}
            />
          )}
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

  function getActions(commit) {
    let actions = [
      { label: 'Edit', onClick: () => setShowEditDialog(true) },
      {
        label: 'Delete',
        onClick: () => setShowDeleteDialog(true),
      },
    ]

    if (!disabled) {
      actions.push({
        label: 'Send Test',
        onClick: () => commit(),
      })
    }

    return actions
  }

  return (
    <Mutation mutation={testCM} client={graphql2Client} variables={{ id }}>
      {commit => (
        <React.Fragment>
          <OtherActions actions={getActions(commit)} />
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
