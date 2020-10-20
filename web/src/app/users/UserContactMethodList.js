import { gql, useQuery } from '@apollo/client'
import React, { useState } from 'react'
import p from 'prop-types'
import FlatList from '../lists/FlatList'
import { Button, Card, CardHeader, Grid, IconButton } from '@material-ui/core'

import { isWidthUp } from '@material-ui/core/withWidth'
import { sortContactMethods } from './util'
import OtherActions from '../util/OtherActions'
import UserContactMethodDeleteDialog from './UserContactMethodDeleteDialog'
import UserContactMethodEditDialog from './UserContactMethodEditDialog'
import { Warning } from '../icons'
import UserContactMethodVerificationDialog from './UserContactMethodVerificationDialog'
import { makeStyles, createStyles } from '@material-ui/core/styles'
import { styles as globalStyles } from '../styles/materialStyles'
import useWidth from '../util/useWidth'
import Spinner from '../loading/components/Spinner'
import { GenericError } from '../error-pages'
import SendTestDialog from './SendTestDialog'

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

const useStyles = makeStyles((theme) => {
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
  const width = useWidth()

  const [showVerifyDialogByID, setShowVerifyDialogByID] = useState(null)
  const [showEditDialogByID, setShowEditDialogByID] = useState(null)
  const [showDeleteDialogByID, setShowDeleteDialogByID] = useState(null)

  const [showSendTestByID, setShowSendTestByID] = useState(null)

  const { loading, error, data } = useQuery(query, {
    variables: {
      id: props.userID,
    },
  })

  if (loading && !data) return <Spinner />
  if (error) return <GenericError error={error.message} />

  const contactMethods = data.user.contactMethods

  const getIcon = (cm) => {
    if (!cm.disabled) return null
    if (props.readOnly) {
      return <Warning message='Contact method disabled' />
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
        <Warning message='Contact method disabled' />
      </IconButton>
    )
  }

  function getActionMenuItems(cm) {
    const actions = [
      { label: 'Edit', onClick: () => setShowEditDialogByID(cm.id) },
      {
        label: 'Delete',
        onClick: () => setShowDeleteDialogByID(cm.id),
      },
    ]

    if (!cm.disabled) {
      actions.push({
        label: 'Send Test',
        onClick: () => setShowSendTestByID(cm.id),
      })
    } else {
      actions.push({
        label: 'Reactivate',
        onClick: () => setShowVerifyDialogByID(cm.id),
      })
    }
    return actions
  }

  function getSecondaryAction(cm) {
    return (
      <Grid container spacing={2} className={classes.actionGrid}>
        {cm.disabled && !props.readOnly && isWidthUp('md', width) && (
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
          items={sortContactMethods(contactMethods).map((cm) => ({
            title: `${cm.name} (${cm.type})${cm.disabled ? ' - Disabled' : ''}`,
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
        {showSendTestByID && (
          <SendTestDialog
            messageID={showSendTestByID}
            onClose={() => setShowSendTestByID(null)}
          />
        )}
      </Card>
    </Grid>
  )
}

UserContactMethodList.propTypes = {
  userID: p.string.isRequired,
  readOnly: p.bool,
}
