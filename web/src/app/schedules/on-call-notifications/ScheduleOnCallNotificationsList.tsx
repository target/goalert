import React, { useState } from 'react'
import { Button, Grid, Card } from '@mui/material'
import Avatar from '@mui/material/Avatar'

import FlatList from '../../lists/FlatList'
import OtherActions from '../../util/OtherActions'
import { SlackBW, WebhookBW } from '../../icons/components/Icons'
import { useOnCallRulesData } from './hooks'
import { onCallRuleSummary } from './util'
import ScheduleOnCallNotificationsCreateDialog from './ScheduleOnCallNotificationsCreateDialog'
import ScheduleOnCallNotificationsDeleteDialog from './ScheduleOnCallNotificationsDeleteDialog'
import CreateFAB from '../../lists/CreateFAB'
import ScheduleOnCallNotificationsEditDialog from './ScheduleOnCallNotificationsEditDialog'
import { useIsWidthDown } from '../../util/useWidth'
import { Add } from '@mui/icons-material'
import { useExpFlag } from '../../util/useExpFlag'
import ScheduleOnCallNotificationsListDest from './ScheduleOnCallNotificationsListDest'

export type ScheduleOnCallNotificationsListProps = {
  scheduleID: string
}

function getChannelIcon(targetType: string): JSX.Element {
  if (targetType === 'slackUserGroup' || targetType === 'slackChannel') {
    return <SlackBW />
  }
  if (targetType === 'chanWebhook') {
    return <WebhookBW />
  }
  return <div />
}

function ScheduleOnCallNotificationsList({
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
                q.fetching
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
                  icon: <Avatar>{getChannelIcon(rule.target.type)}</Avatar>,
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

// Since we are using a feature flag, and cannot conditionally update the router (which always renders the default export of this file),
// we need to create a component that will conditionally render the destination version based on the feature flag.
export default function ScheduleOnCallNotificationsListSwitcher(
  props: ScheduleOnCallNotificationsListProps,
): JSX.Element {
  const hasDestTypesFlag = useExpFlag('dest-types')

  if (hasDestTypesFlag) {
    return <ScheduleOnCallNotificationsListDest {...props} />
  }

  return <ScheduleOnCallNotificationsList {...props} />
}
