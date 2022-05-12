import React, { useState, ReactNode } from 'react'
import Query from '../util/Query'
import { gql, QueryResult } from '@apollo/client'
import FlatList, { FlatListListItem } from '../lists/FlatList'
import { Grid, Card, CardHeader, IconButton } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import { formatNotificationRule, sortNotificationRules } from './util'
import { Delete } from '@mui/icons-material'
import UserNotificationRuleDeleteDialog from './UserNotificationRuleDeleteDialog'
import { styles as globalStyles } from '../styles/materialStyles'

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

const useStyles = makeStyles((theme) => {
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
  const [deleteID, setDeleteID] = useState(null)

  function renderList(notificationRules: FlatListListItem[]): ReactNode {
    return (
      <Grid item xs={12}>
        <Card>
          <CardHeader
            className={classes.cardHeader}
            titleTypographyProps={{ component: 'h2', variant: 'h5' }}
            title='Notification Rules'
          />
          <FlatList
            data-cy='notification-rules'
            items={sortNotificationRules(notificationRules).map((nr) => ({
              title: formatNotificationRule(nr.delayMinutes, nr.contactMethod),
              secondaryAction: props.readOnly ? null : (
                <IconButton
                  aria-label='Delete notification rule'
                  onClick={() => setDeleteID(nr.id)}
                  size='large'
                >
                  <Delete />
                </IconButton>
              ),
            }))}
            emptyMessage='No notification rules'
          />
        </Card>
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
