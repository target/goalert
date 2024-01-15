import React, { useState, ReactNode } from 'react'
import { gql, useQuery } from 'urql'
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
import { GenericError, ObjectNotFound } from '../error-pages'
import SendTestDialog from './SendTestDialog'
import AppLink from '../util/AppLink'
import { styles as globalStyles } from '../styles/materialStyles'
import { UserContactMethod } from '../../schema'
import UserContactMethodCreateDialog from './UserContactMethodCreateDialog'
import { useSessionInfo } from '../util/RequireConfig'
import { useExpFlag } from '../util/useExpFlag'
import UserContactMethodEditDialogDest from './UserContactMethodEditDialogDest'

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
        pending
      }
    }
  }
`

interface ListItemAction {
  label: string
  onClick: () => void
  disabled?: boolean
  tooltip?: string
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
): ReactNode {
  const classes = useStyles()
  const mobile = useIsWidthDown('md')

  const isDestTypesSet = useExpFlag('dest-types')

  const [showAddDialog, setShowAddDialog] = useState(false)
  const [showVerifyDialogByID, setShowVerifyDialogByID] = useState('')
  const [showEditDialogByID, setShowEditDialogByID] = useState('')
  const [showDeleteDialogByID, setShowDeleteDialogByID] = useState('')
  const [showSendTestByID, setShowSendTestByID] = useState('')

  const [{ error, data }] = useQuery({
    query,
    variables: {
      id: props.userID,
    },
  })

  const { userID: currentUserID } = useSessionInfo()
  const isCurrentUser = props.userID === currentUserID

  if (!data?.user) return <ObjectNotFound type='user' />
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
      {
        label: 'Edit',
        onClick: () => setShowEditDialogByID(cm.id),
        disabled: false,
        tooltip: '',
      },
      {
        label: 'Delete',
        onClick: () => setShowDeleteDialogByID(cm.id),
        disabled: false,
        tooltip: '',
      },
    ]

    // disable send test and reactivate if not current user
    if (!cm.disabled) {
      actions.push({
        label: 'Send Test',
        onClick: () => setShowSendTestByID(cm.id),
        disabled: !isCurrentUser,
        tooltip: !isCurrentUser
          ? 'Send Test only available for your own contact methods'
          : '',
      })
    } else {
      actions.push({
        label: 'Reactivate',
        onClick: () => setShowVerifyDialogByID(cm.id),
        disabled: !isCurrentUser,
        tooltip: !isCurrentUser
          ? 'Reactivate only available for your own contact methods'
          : '',
      })
    }
    return actions
  }

  function getSecondaryAction(cm: UserContactMethod): JSX.Element {
    return (
      <Grid container spacing={2} alignItems='center' wrap='nowrap'>
        {cm.disabled && !props.readOnly && !mobile && isCurrentUser && (
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
    let cmText = cm.formattedValue
    if (cm.pending) {
      cmText = `${cm.formattedValue} - this contact method will be automatically deleted if not verified`
    }
    if (cm.type === 'WEBHOOK') {
      return (
        <React.Fragment>
          {`${cmText} (`}
          <AppLink to='/docs'>docs</AppLink>)
        </React.Fragment>
      )
    }

    return cmText
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
              <Button
                title='Create Contact Method'
                variant='contained'
                onClick={() => setShowAddDialog(true)}
                startIcon={<Add />}
              >
                Create Method
              </Button>
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
        {showEditDialogByID &&
          (isDestTypesSet ? (
            <UserContactMethodEditDialogDest
              contactMethodID={showEditDialogByID}
              onClose={() => setShowEditDialogByID('')}
            />
          ) : (
            <UserContactMethodEditDialog
              contactMethodID={showEditDialogByID}
              onClose={() => setShowEditDialogByID('')}
            />
          ))}
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
