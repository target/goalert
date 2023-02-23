import React, { useState } from 'react'
import { Button, Grid, Card } from '@mui/material'
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
import { useIsWidthDown } from '../../util/useWidth'
import { Add } from '@mui/icons-material'

export type ScheduleOnCallNotificationsListProps = {
  scheduleID: string
}

export default function ScheduleOnCallNotificationsList({
  scheduleID,
}: ScheduleOnCallNotificationsListProps): JSX.Element {
  const [createRule, setCreateRule] = useState(false)
  const [editRuleID, setEditRuleID] = useState('')
  const [deleteRuleID, setDeleteRuleID] = useState('')
  const isMobile = useIsWidthDown('md')

  const { q, zone, rules } = useOnCallRulesData(scheduleID)

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
              headerAction={
                isMobile ? undefined : (
                  <Button
                    variant='contained'
                    onClick={() => setCreateRule(true)}
                    startIcon={<Add />}
                  >
                    Create Notification Rule
                  </Button>
                )
              }
              items={rules.map((rule) => {
                return {
                  icon:
                    rule.target.type === 'slackChannel' ? (
                      <Avatar>
                        <SlackBW />{' '}
                      </Avatar>
                    ) : null,
                  title: rule.target.name ?? undefined,
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
      {isMobile && (
        <CreateFAB
          onClick={() => setCreateRule(true)}
          title='Create Notification Rule'
        />
      )}
      {createRule && (
        <ScheduleOnCallNotificationsCreateDialog
          scheduleID={scheduleID}
          onClose={() => setCreateRule(false)}
        />
      )}
      {editRuleID && (
        <ScheduleOnCallNotificationsEditDialog
          scheduleID={scheduleID}
          ruleID={editRuleID}
          onClose={() => setEditRuleID('')}
        />
      )}
      {deleteRuleID && (
        <ScheduleOnCallNotificationsDeleteDialog
          scheduleID={scheduleID}
          ruleID={deleteRuleID}
          onClose={() => setDeleteRuleID('')}
        />
      )}
    </React.Fragment>
  )
}
