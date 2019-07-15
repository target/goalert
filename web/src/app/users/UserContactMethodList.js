import React, { useState, useEffect } from 'react'
import p from 'prop-types'
import Query from '../util/Query'
import gql from 'graphql-tag'
import FlatList from '../lists/FlatList'
import { Button, Card, CardHeader, Grid, IconButton } from '@material-ui/core'
import { formatCMValue, sortContactMethods } from './util'
import OtherActions from '../util/OtherActions'
import UserContactMethodDeleteDialog from './UserContactMethodDeleteDialog'
import UserContactMethodEditDialog from './UserContactMethodEditDialog'
import ListItem from '@material-ui/core/ListItem'
import ListItemText from '@material-ui/core/ListItemText'
import { Config } from '../util/RequireConfig'
import { Warning } from '../icons'
import UserContactMethodVerificationDialog from './UserContactMethodVerificationDialog'
import { makeStyles } from '@material-ui/core/styles'
import { useMutation } from '@apollo/react-hooks'

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

const useStyles = makeStyles({
  actionGrid: {
    display: 'flex',
    alignItems: 'center',
  },
})

export default function UserContactMethodList(props) {
  const classes = useStyles()

  const [showVerifyDialogByID, setShowVerifyDialogByID] = useState(null)
  const [showEditDialogByID, setShowEditDialogByID] = useState(null)
  const [showDeleteDialogByID, setShowDeleteDialogByID] = useState(null)
  const [testID, setTestID] = useState(null)

  const [sendTest] = useMutation(testCM, {
    variables: {
      id: testID,
    },
  })

  // send test to CM after testID is set
  useEffect(() => {
    if (testID) {
      sendTest().catch(err => console.error(err.message))
    }
  }, [testID])

  const getIcon = cm => {
    if (cm.disabled && props.readOnly) {
      return <Warning title='Contact method disabled' />
    } else if (cm.disabled && !props.readOnly) {
      return (
        <IconButton
          aria-label='Reactivate contact method'
          onClick={() => setShowVerifyDialogByID(cm.id)}
          variant='contained'
          color='primary'
          disabled={props.readOnly}
        >
          <Warning title='Contact method disabled' />
        </IconButton>
      )
    }
  }

  function getActionMenuItems(cm) {
    let actions = [
      { label: 'Edit', onClick: () => setShowEditDialogByID(cm.id) },
      {
        label: 'Delete',
        onClick: () => setShowDeleteDialogByID(cm.id),
      },
    ]

    if (!cm.disabled) {
      actions.push({
        label: 'Send Test',
        onClick: () => setTestID(cm.id),
      })
    }

    return actions
  }

  function getSecondaryAction(cm) {
    return (
      <Grid container spacing={2} className={classes.actionGrid}>
        {cm.disabled && !props.readOnly && (
          <Grid item>
            <Button
              aria-label='Reactivate contact method'
              onClick={() => setShowVerifyDialogByID(cm.id)}
              variant='contained'
              color='primary'
            >
              Reactivate
            </Button>
          </Grid>
        )}
        {!props.readOnly && (
          <Grid item>
            <OtherActions actions={getActionMenuItems(cm)} />
          </Grid>
        )}
      </Grid>
    )
  }

  function renderList(contactMethods) {
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
          {showVerifyDialogByID && (
            <UserContactMethodVerificationDialog
              contactMethodID={showVerifyDialogByID}
              onClose={() => setShowVerifyDialogByID(null)}
            />
          )}
          {showEditDialogByID && (
            <UserContactMethodEditDialog
              contactMethodID={showEditDialogByID}
              onClose={() => setShowEditDialogByID(null)}
            />
          )}
          {showDeleteDialogByID && (
            <UserContactMethodDeleteDialog
              contactMethodID={showDeleteDialogByID}
              onClose={() => setShowDeleteDialogByID(null)}
            />
          )}
          <Config>
            {cfg =>
              !props.readOnly &&
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

  return (
    <Query
      query={query}
      variables={{ id: props.userID }}
      render={({ data }) => renderList(data.user.contactMethods)}
    />
  )
}

UserContactMethodList.propTypes = {
  userID: p.string.isRequired,
  readOnly: p.bool,
}
