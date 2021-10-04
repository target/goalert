import React, { useState } from 'react'
import { Grid, Card } from '@mui/material'
import Avatar from '@mui/material/Avatar'

import FlatList from '../../lists/FlatList'
import OtherActions from '../../util/OtherActions'
import { SlackBW } from '../../icons/components/Icons'
import { useOnCallRulesData } from './hooks'
import { onCallRuleSummary } from './util'
import ScheduleOnCallNotificationsCreateDialog from './ScheduleOnCallNotificationsCreateDialog'
import ScheduleOnCallNotificationsDeleteDialog from './ScheduleOnCallNotificationsDeleteDialog'
import CreateFAB from '../../lists/CreateFAB'
import ScheduleOnCallNotificationsEditDialog from './ScheduleOnCallNotificationsEditDialog'

interface ScheduleOnCallNotificationsListProps {
  scheduleID: string
}

export default function ScheduleOnCallNotificationsList(
  props: ScheduleOnCallNotificationsListProps,
): JSX.Element {
  const [createRule, setCreateRule] = useState(false)
  const [editRuleID, setEditRuleID] = useState('')
  const [deleteRuleID, setDeleteRuleID] = useState('')
  const { q, zone, rules } = useOnCallRulesData(props.scheduleID)

  return (
    <React.Fragment>
      <Grid container spacing={2}>
        <Grid item xs={12}>
          <Card>
            <FlatList
              headerNote={zone ? `Showing times for schedule in ${zone}.` : ''}
              emptyMessage={
                q.loading
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
                  subText: 'Notifies ' + onCallRuleSummary(zone, rule),
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
