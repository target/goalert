import React, { useState } from 'react'
import p from 'prop-types'
import Query from '../util/Query'
import gql from 'graphql-tag'
import FlatList from '../lists/FlatList'
import { Button, Card, CardHeader, Grid, IconButton } from '@material-ui/core'
import { sortContactMethods } from './util'
import OtherActions from '../util/OtherActions'
import UserContactMethodDeleteDialog from './UserContactMethodDeleteDialog'
import UserContactMethodEditDialog from './UserContactMethodEditDialog'
import ListItem from '@material-ui/core/ListItem'
import ListItemText from '@material-ui/core/ListItemText'
import { Config } from '../util/RequireConfig'
import { Warning } from '../icons'
import UserContactMethodVerificationDialog from './UserContactMethodVerificationDialog'
import { makeStyles, createStyles } from '@material-ui/core/styles'
import { useMutation } from '@apollo/react-hooks'
import { styles as globalStyles } from '../styles/materialStyles'

const query = gql`
  query cmList($id: ID!) {
    user(id: $id) {
      id
      contactMethods {
        id
        name
        type
        value
        formattedValue
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

const useStyles = makeStyles(theme => {
  const { cardHeader } = globalStyles(theme)

  return createStyles({
    actionGrid: {
      display: 'flex',
      alignItems: 'center',
    },
    cardHeader,
  })
})

export default function UserContactMethodList(props) {
  const classes = useStyles()

  const [showVerifyDialogByID, setShowVerifyDialogByID] = useState(null)
  const [showEditDialogByID, setShowEditDialogByID] = useState(null)
  const [showDeleteDialogByID, setShowDeleteDialogByID] = useState(null)

  const [sendTest] = useMutation(testCM)

  const getIcon = cm => {
    if (!cm.disabled) return null
    if (props.readOnly) {
      return <Warning title='Contact method disabled' />
    }

    return (
      <IconButton
        data-cy='cm-disabled'
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
        // todo: show dialog with error if test message fails to send
        onClick: () =>
          sendTest({
            variables: {
              id: cm.id,
            },
          }),
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
          <CardHeader
            className={classes.cardHeader}
            component='h3'
            title='Contact Methods'
          />
          <FlatList
            data-cy='contact-methods'
            items={sortContactMethods(contactMethods).map(cm => ({
              title: `${cm.name} (${cm.type})${
                cm.disabled ? ' - Disabled' : ''
              }`,
              subText: cm.formattedValue,
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
