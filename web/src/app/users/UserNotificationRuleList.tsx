import React, { useState, ReactNode } from 'react'
import { gql, useQuery } from 'urql'
import {
  Button,
  Card,
  CardHeader,
  Grid,
  IconButton,
  Theme,
} from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import { Add, Delete } from '@mui/icons-material'
import FlatList from '../lists/FlatList'
import { formatNotificationRule, sortNotificationRules } from './util'
import UserNotificationRuleDeleteDialog from './UserNotificationRuleDeleteDialog'
import { styles as globalStyles } from '../styles/materialStyles'
import UserNotificationRuleCreateDialog from './UserNotificationRuleCreateDialog'
import { useIsWidthDown } from '../util/useWidth'
import { User } from '../../schema'
import { useSuspenseContext } from './UserDetails'
import { ObjectNotFound, GenericError } from '../error-pages'

const query = gql`
  query nrList($id: ID!) {
    user(id: $id) {
      id
      contactMethods {
        id
      }
      notificationRules {
        id
        delayMinutes
        contactMethod {
          id
          type
          name
          value
          formattedValue
        }
      }
    }
  }
`

const useStyles = makeStyles((theme: Theme) => {
  const { cardHeader } = globalStyles(theme)
  return {
    cardHeader,
  }
})

export default function UserNotificationRuleList(props: {
  userID: string
  readOnly: boolean
}): ReactNode {
  const classes = useStyles()
  const mobile = useIsWidthDown('md')
  const [showAddDialog, setShowAddDialog] = useState(false)
  const [deleteID, setDeleteID] = useState(null)

  const [{ data, error, fetching }] = useQuery({
    query,
    variables: { id: props.userID },
    context: useSuspenseContext(),
  })

  if (fetching && !data) return null
  if (data && !data.user)
    return <ObjectNotFound type='notifcation rules list' />
  if (error) return <GenericError error={error.message} />

  const { user }: { user: User } = data

  return (
    <Grid item xs={12}>
      <Card>
        <CardHeader
          className={classes.cardHeader}
          titleTypographyProps={{ component: 'h2', variant: 'h5' }}
          title='Notification Rules'
          action={
            !mobile ? (
              <Button
                title='Add Notification Rule'
                variant='contained'
                onClick={() => setShowAddDialog(true)}
                startIcon={<Add />}
                disabled={user.contactMethods.length === 0}
              >
                Add Rule
              </Button>
            ) : null
          }
        />
        <FlatList
          data-cy='notification-rules'
          items={sortNotificationRules(user.notificationRules).map((nr) => ({
            title: formatNotificationRule(nr.delayMinutes, nr.contactMethod),
            secondaryAction: props.readOnly ? null : (
              <IconButton
                aria-label='Delete notification rule'
                onClick={() => setDeleteID(nr.id)}
                color='secondary'
              >
                <Delete />
              </IconButton>
            ),
          }))}
          emptyMessage='No notification rules'
        />
      </Card>
      {showAddDialog && (
        <UserNotificationRuleCreateDialog
          userID={props.userID}
          onClose={() => setShowAddDialog(false)}
        />
      )}
      {deleteID && (
        <UserNotificationRuleDeleteDialog
          ruleID={deleteID}
          onClose={() => setDeleteID(null)}
        />
      )}
    </Grid>
  )
}
