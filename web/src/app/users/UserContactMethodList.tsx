import React, { useState } from 'react'
import { gql, useQuery } from '@apollo/client'
import FlatList from '../lists/FlatList'
import { Button, Card, CardHeader, Grid, IconButton } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import { Theme } from '@mui/material/styles'
import { Add } from '@mui/icons-material'
import { sortContactMethods } from './util'
import OtherActions from '../util/OtherActions'
import UserContactMethodDeleteDialog from './UserContactMethodDeleteDialog'
import UserContactMethodEditDialog from './UserContactMethodEditDialog'
import { Warning } from '../icons'
import UserContactMethodVerificationDialog from './UserContactMethodVerificationDialog'
import { useIsWidthDown } from '../util/useWidth'
import Spinner from '../loading/components/Spinner'
import { GenericError, ObjectNotFound } from '../error-pages'
import SendTestDialog from './SendTestDialog'
import AppLink from '../util/AppLink'
import { styles as globalStyles } from '../styles/materialStyles'
import { UserContactMethod } from '../../schema'
import UserContactMethodCreateDialog from './UserContactMethodCreateDialog'
import { useExpFlag } from '../util/useExpFlag'

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

const useStyles = makeStyles((theme: Theme) => ({
  cardHeader: globalStyles(theme).cardHeader,
}))

export default function UserContactMethodList(
  props: UserContactMethodListProps,
): JSX.Element {
  const classes = useStyles()
  const mobile = useIsWidthDown('md')

  const [showAddDialog, setShowAddDialog] = useState(false)
  const [showVerifyDialogByID, setShowVerifyDialogByID] = useState('')
  const [showEditDialogByID, setShowEditDialogByID] = useState('')
  const [showDeleteDialogByID, setShowDeleteDialogByID] = useState('')
  const [showSendTestByID, setShowSendTestByID] = useState('')
  const hasSlackDM = useExpFlag('slack-dm')

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
        disabled={props.readOnly}
        size='large'
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

    // don't show send test for slack DMs if disabled
    if (cm.type === 'SLACK_DM' && !hasSlackDM) return actions

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
        {cm.disabled && !props.readOnly && !mobile && (
          <Grid item>
            <Button
              aria-label='Reactivate contact method'
              onClick={() => setShowVerifyDialogByID(cm.id)}
              variant='contained'
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
          action={
            !mobile ? (
              <IconButton
                title='Add contact method'
                onClick={() => setShowAddDialog(true)}
                size='large'
              >
                <Add fontSize='large' />
              </IconButton>
            ) : null
          }
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
        {showAddDialog && (
          <UserContactMethodCreateDialog
            userID={props.userID}
            onClose={(contactMethodID = '') => {
              setShowAddDialog(false)
              setShowVerifyDialogByID(contactMethodID)
            }}
          />
        )}
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
