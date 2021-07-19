import React, { useState } from 'react'
import { gql, useQuery } from '@apollo/client'
import FlatList from '../lists/FlatList'
import {
  Button,
  Card,
  CardHeader,
  Grid,
  IconButton,
  makeStyles,
} from '@material-ui/core'

import { isWidthUp } from '@material-ui/core/withWidth'
import { sortContactMethods } from './util'
import OtherActions from '../util/OtherActions'
import UserContactMethodDeleteDialog from './UserContactMethodDeleteDialog'
import UserContactMethodEditDialog from './UserContactMethodEditDialog'
import { Warning } from '../icons'
import UserContactMethodVerificationDialog from './UserContactMethodVerificationDialog'
import useWidth from '../util/useWidth'
import Spinner from '../loading/components/Spinner'
import { GenericError, ObjectNotFound } from '../error-pages'
import SendTestDialog from './SendTestDialog'
import AppLink from '../util/AppLink'
import { styles as globalStyles } from '../styles/materialStyles'
import { UserContactMethod } from '../../schema'

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

interface ListItemAction {
  label: string
  onClick: () => void
}

interface UserContactMethodListProps {
  userID: string
  readOnly?: boolean
}

const useStyles = makeStyles((theme) => ({
  cardHeader: globalStyles(theme).cardHeader,
}))

export default function UserContactMethodList(
  props: UserContactMethodListProps,
): JSX.Element {
  const classes = useStyles()
  const width = useWidth()
  const [showVerifyDialogByID, setShowVerifyDialogByID] = useState('')
  const [showEditDialogByID, setShowEditDialogByID] = useState('')
  const [showDeleteDialogByID, setShowDeleteDialogByID] = useState('')
  const [showSendTestByID, setShowSendTestByID] = useState('')

  const { loading, error, data } = useQuery(query, {
    variables: {
      id: props.userID,
    },
  })

  if (loading && !data) return <Spinner />
  if (data && !data.user) return <ObjectNotFound type='user' />
  if (error) return <GenericError error={error.message} />

  const contactMethods = data.user.contactMethods

  const getIcon = (cm: UserContactMethod): JSX.Element | null => {
    if (!cm.disabled) return null
    if (props.readOnly) {
      return <Warning message='Contact method disabled' />
    }

    return (
      <IconButton
        data-cy='cm-disabled'
        aria-label='Reactivate contact method'
        onClick={() => setShowVerifyDialogByID(cm.id)}
        color='primary'
        disabled={props.readOnly}
      >
        <Warning message='Contact method disabled' />
      </IconButton>
    )
  }

  function getActionMenuItems(cm: UserContactMethod): ListItemAction[] {
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

  function getSecondaryAction(cm: UserContactMethod): JSX.Element {
    return (
      <Grid container spacing={2} alignItems='center' wrap='nowrap'>
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

  function getSubText(cm: UserContactMethod): JSX.Element | string {
    if (cm.type === 'WEBHOOK') {
      return (
        <React.Fragment>
          {`${cm.formattedValue} (`}
          <AppLink to='/docs'>docs</AppLink>)
        </React.Fragment>
      )
    }

    return cm.formattedValue
  }

  return (
    <Grid item xs={12}>
      <Card>
        <CardHeader
          className={classes.cardHeader}
          titleTypographyProps={{ component: 'h2', variant: 'h5' }}
          title='Contact Methods'
        />
        <FlatList
          data-cy='contact-methods'
          items={sortContactMethods(contactMethods).map((cm) => ({
            title: `${cm.name} (${cm.type})${cm.disabled ? ' - Disabled' : ''}`,
            subText: getSubText(cm),
            secondaryAction: getSecondaryAction(cm),
            icon: getIcon(cm),
          }))}
          emptyMessage='No contact methods'
        />
        {showVerifyDialogByID && (
          <UserContactMethodVerificationDialog
            contactMethodID={showVerifyDialogByID}
            onClose={() => setShowVerifyDialogByID('')}
          />
        )}
        {showEditDialogByID && (
          <UserContactMethodEditDialog
            contactMethodID={showEditDialogByID}
            onClose={() => setShowEditDialogByID('')}
          />
        )}
        {showDeleteDialogByID && (
          <UserContactMethodDeleteDialog
            contactMethodID={showDeleteDialogByID}
            onClose={() => setShowDeleteDialogByID('')}
          />
        )}
        {showSendTestByID && (
          <SendTestDialog
            messageID={showSendTestByID}
            onClose={() => setShowSendTestByID('')}
          />
        )}
      </Card>
    </Grid>
  )
}
