import React, { useState } from 'react'
import p from 'prop-types'
import Query from '../util/Query'
import { gql } from '@apollo/client'
import FlatList from '../lists/FlatList'
import {
  Grid,
  Card,
  CardHeader,
  IconButton,
  makeStyles,
} from '@material-ui/core'
import { formatNotificationRule, sortNotificationRules } from './util'
import { Delete } from '@material-ui/icons'
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

export default function UserNotificationRuleList({ userID, readOnly }) {
  const classes = useStyles()
  const [deleteId, setDeleteId] = useState(null)

  function renderList(notificationRules) {
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
              secondaryAction: readOnly ? null : (
                <IconButton
                  aria-label='Delete notification rule'
                  onClick={() => setDeleteId(nr.id)}
                >
                  <Delete />
                </IconButton>
              ),
            }))}
            emptyMessage='No notification rules'
          />
        </Card>
        {deleteId && (
          <UserNotificationRuleDeleteDialog
            ruleID={deleteId}
            onClose={() => setDeleteId(null)}
          />
        )}
      </Grid>
    )
  }
  return (
    <Query
      query={query}
      variables={{ id: userID }}
      render={({ data }) => renderList(data.user.notificationRules)}
    />
  )
}

UserNotificationRuleList.propTypes = {
  userID: p.string.isRequired,
  readOnly: p.bool,
}
