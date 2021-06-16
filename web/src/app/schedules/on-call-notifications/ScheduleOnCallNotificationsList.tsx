import React, { useState } from 'react'
import { Grid, Card } from '@material-ui/core'
import Avatar from '@material-ui/core/Avatar'

import FlatList from '../../lists/FlatList'
import OtherActions from '../../util/OtherActions'
import { SlackBW } from '../../icons/components/Icons'
import { Rule } from './util'
import ScheduleOnCallNotificationsCreateDialog from './ScheduleOnCallNotificationsCreateDialog'
import ScheduleOnCallNotificationsDeleteDialog from './ScheduleOnCallNotificationsDeleteDialog'
import { gql, useQuery } from '@apollo/client'
import { OnCallNotificationRule } from '../../../schema'
import { DateTime } from 'luxon'
import { weekdaySummary } from '../util'
import CreateFAB from '../../lists/CreateFAB'
import ScheduleOnCallNotificationsEditDialog from './ScheduleOnCallNotificationsEditDialog'

type ScheduleOnCallNotificationsListProps = {
  scheduleID: string
}

const query = gql`
  query ($id: ID!) {
    schedule(id: $id) {
      id
      timeZone
      onCallNotificationRules {
        id
        target {
          id
          type
          name
        }
        time
        weekdayFilter
      }
    }
  }
`

function ruleSummary(scheduleZone: string, r: Rule): string {
  const prefix = `Notifies`
  if (!r.time) {
    return `${prefix} when on-call changes.`
  }

  const dt = DateTime.fromFormat(r.time, 'HH:mm', {
    zone: scheduleZone,
  })

  const timeStr = dt.toLocaleString(DateTime.TIME_SIMPLE)
  const localStr = dt.setZone('local').toLocaleString(DateTime.TIME_SIMPLE)

  const summary = `${prefix} ${weekdaySummary(r.weekdayFilter)} at ${timeStr}`
  if (timeStr === localStr) {
    return summary
  }

  return summary + ` (${localStr} ${dt.setZone('local').toFormat('ZZZZ')})`
}

export default function ScheduleOnCallNotificationsList(
  props: ScheduleOnCallNotificationsListProps,
): JSX.Element {
  const [createRule, setCreateRule] = useState(false)
  const [editRuleID, setEditRuleID] = useState('')
  const [deleteRuleID, setDeleteRuleID] = useState('')
  const { data, loading, error } = useQuery(query, {
    variables: { id: props.scheduleID },
  })

  const rules: OnCallNotificationRule[] = (
    data?.schedule?.onCallNotificationRules ?? []
  ).filter((r) => r) // remove any invalid/null rules
  const tz = data?.schedule?.timeZone

  return (
    <React.Fragment>
      <Grid container spacing={2}>
        <Grid item xs={12}>
          <Card>
            <FlatList
              headerNote={tz ? `Showing times for schedule in ${tz}.` : ''}
              emptyMessage={
                !data && loading
                  ? 'Loading notification rules...'
                  : 'No notification rules.'
              }
              items={rules.map((rule) => {
                return {
                  icon:
                    rule.target.type === 'slackChannel' ? (
                      <Avatar>
                        <SlackBW />{' '}
                      </Avatar>
                    ) : null,
                  title: rule.target.name,
                  subText: ruleSummary(data.schedule.timeZone, rule),
                  secondaryAction: (
                    <OtherActions
                      actions={[
                        {
                          label: 'Edit',
                          onClick: () => setEditRuleID(rule.id),
                        },
                        {
                          label: 'Delete',
                          onClick: () => setDeleteRuleID(rule.id),
                        },
                      ]}
                    />
                  ),
                }
              })}
            />
          </Card>
        </Grid>
      </Grid>
      {createRule && (
        <ScheduleOnCallNotificationsCreateDialog
          scheduleID={props.scheduleID}
          onClose={() => setCreateRule(false)}
        />
      )}
      {editRuleID && (
        <ScheduleOnCallNotificationsEditDialog
          scheduleID={props.scheduleID}
          ruleID={editRuleID}
          onClose={() => setEditRuleID('')}
        />
      )}
      {deleteRuleID && (
        <ScheduleOnCallNotificationsDeleteDialog
          scheduleID={props.scheduleID}
          ruleID={deleteRuleID}
          onClose={() => setDeleteRuleID('')}
        />
      )}
      <CreateFAB
        onClick={() => setCreateRule(true)}
        title='Create Notification Rule'
      />
    </React.Fragment>
  )
}
