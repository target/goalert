import React from 'react'
import p from 'prop-types'
import Query from '../util/Query'
import { gql } from '@apollo/client'
import FlatList from '../lists/FlatList'
import {
  Grid,
  Card,
  CardHeader,
  IconButton,
  withStyles,
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

const styles = (theme) => {
  const { cardHeader } = globalStyles(theme)

  return {
    cardHeader,
  }
}

@withStyles(styles)
export default class UserNotificationRuleList extends React.PureComponent {
  static propTypes = {
    userID: p.string.isRequired,
    readOnly: p.bool,
  }

  state = {
    delete: null,
  }

  render() {
    return (
      <Query
        query={query}
        variables={{ id: this.props.userID }}
        render={({ data }) => this.renderList(data.user.notificationRules)}
      />
    )
  }

  renderList(notificationRules) {
    const { classes } = this.props
    return (
      <Grid item xs={12}>
        <Card>
          <CardHeader
            className={classes.cardHeader}
            component='h3'
            title='Notification Rules'
          />
          <FlatList
            data-cy='notification-rules'
            items={sortNotificationRules(notificationRules).map((nr) => ({
              title: formatNotificationRule(nr.delayMinutes, nr.contactMethod),
              secondaryAction: this.props.readOnly ? null : (
                <IconButton
                  aria-label='Delete notification rule'
                  onClick={() => this.setState({ delete: nr.id })}
                >
                  <Delete />
                </IconButton>
              ),
            }))}
            emptyMessage='No notification rules'
          />
        </Card>
        {this.state.delete && (
          <UserNotificationRuleDeleteDialog
            ruleID={this.state.delete}
            onClose={() => this.setState({ delete: null })}
          />
        )}
      </Grid>
    )
  }
}
