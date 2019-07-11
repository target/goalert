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

export default class UserContactMethodList extends React.PureComponent {
  static propTypes = {
    userID: p.string.isRequired,
    readOnly: p.bool,
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
          <ReactivateButton
            ButtonComponent={IconButton}
            buttonChild={<Warning title='Contact method disabled' />}
            contactMethodID={cm.id}
            disabled={readOnly}
          />
        )
      }
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
              secondaryAction: readOnly ? null : <Actions contactMethod={cm} />,
              tertiaryAction:
                cm.disabled && !readOnly ? (
                  <ReactivateButton
                    ButtonComponent={Button}
                    buttonChild='Reactivate'
                    contactMethodID={cm.id}
                  />
                ) : null,
              icon: getIcon(cm),
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

function ReactivateButton(props) {
  const { contactMethodID, ButtonComponent, buttonChild, ...rest } = props
  const [showVerifyDialog, setShowVerifyDialog] = useState(false)

  return (
    <React.Fragment>
      <ButtonComponent
        aria-label='Reactivate contact method'
        onClick={() => setShowVerifyDialog(true)}
        variant='contained'
        color='primary'
        {...rest}
      >
        {buttonChild}
      </ButtonComponent>
      {showVerifyDialog && (
        <UserContactMethodVerificationDialog
          contactMethodID={contactMethodID}
          onClose={() => setShowVerifyDialog(false)}
        />
      )}
    </React.Fragment>
  )
}

ReactivateButton.propTypes = {
  contactMethodID: p.string.isRequired,
  ButtonComponent: p.object.isRequired,
  buttonChild: p.node.isRequired,
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
