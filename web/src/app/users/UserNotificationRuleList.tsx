import React, { useState, ReactNode } from 'react'
import { gql, QueryResult } from '@apollo/client'
import { Grid, Card, CardHeader, IconButton, Theme } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import { Add, Delete } from '@mui/icons-material'
import Query from '../util/Query'
import FlatList, { FlatListListItem } from '../lists/FlatList'
import { formatNotificationRule, sortNotificationRules } from './util'
import UserNotificationRuleDeleteDialog from './UserNotificationRuleDeleteDialog'
import { styles as globalStyles } from '../styles/materialStyles'
import UserNotificationRuleCreateDialog from './UserNotificationRuleCreateDialog'
import { useIsWidthDown } from '../util/useWidth'

const query = gql`
  query nrList($id: ID!) {
    user(id: $id) {
      id
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
}): JSX.Element {
  const classes = useStyles()
  const mobile = useIsWidthDown('md')
  const [showAddDialog, setShowAddDialog] = useState(false)
  const [deleteID, setDeleteID] = useState(null)

  function renderList(notificationRules: FlatListListItem[]): ReactNode {
    return (
      <Grid item xs={12}>
        <Card>
          <CardHeader
            className={classes.cardHeader}
            titleTypographyProps={{ component: 'h2', variant: 'h5' }}
            title='Notification Rules'
            action={
              !mobile ? (
                <IconButton
                  aria-label='Add notification rule'
                  onClick={() => setShowAddDialog(true)}
                  size='large'
                >
                  <Add fontSize='large' />
                </IconButton>
              ) : null
            }
          />
          <FlatList
            data-cy='notification-rules'
            items={sortNotificationRules(notificationRules).map((nr) => ({
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
  return (
    <Query
      query={query}
      variables={{ id: props.userID }}
      render={({ data }: QueryResult) =>
        renderList(data.user.notificationRules)
      }
    />
  )
}
