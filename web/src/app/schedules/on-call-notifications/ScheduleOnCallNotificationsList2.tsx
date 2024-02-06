import React, { Suspense, useState } from 'react'
import { Button, Grid, Card } from '@mui/material'
import FlatList from '../../lists/FlatList'
import OtherActions from '../../util/OtherActions'
import { onCallRuleSummary } from './util'
import ScheduleOnCallNotificationsCreateDialog from './ScheduleOnCallNotificationsCreateDialog'
import ScheduleOnCallNotificationsDeleteDialog from './ScheduleOnCallNotificationsDeleteDialog'
import CreateFAB from '../../lists/CreateFAB'
import ScheduleOnCallNotificationsEditDialog from './ScheduleOnCallNotificationsEditDialog'
import { useIsWidthDown } from '../../util/useWidth'
import { Add } from '@mui/icons-material'
import { gql, useQuery } from 'urql'
import { Schedule } from '../../../schema'
import { DestinationAvatar } from '../../util/DestinationAvatar'

export type ScheduleOnCallNotificationsList2Props = {
  scheduleID: string
}

const query = gql`
  query ScheduleNotifications($scheduleID: ID!) {
    schedule(id: $scheduleID) {
      id
      timeZone
      onCallNotificationRules {
        id
        time
        weekdayFilter
        dest {
          display {
            text
            iconURL
            iconAltText
          }
        }
      }
    }
  }
`

export default function ScheduleOnCallNotificationsList2({
  scheduleID,
}: ScheduleOnCallNotificationsList2Props): JSX.Element {
  const [createRule, setCreateRule] = useState(false)
  const [editRuleID, setEditRuleID] = useState('')
  const [deleteRuleID, setDeleteRuleID] = useState('')
  const isMobile = useIsWidthDown('md')

  const [q] = useQuery<{ schedule: Schedule }>({
    query,
    variables: { scheduleID },
  })

  if (q.error || !q.data) {
    return (
      <Grid container spacing={2}>
        <Grid item xs={12}>
          <Card>
            <FlatList
              headerNote='Error loading notification rules'
              emptyMessage={q.error?.message}
              items={[]}
            />
          </Card>
        </Grid>
      </Grid>
    )
  }

  const schedule = q.data.schedule
  const timeZone = schedule.timeZone

  return (
    <React.Fragment>
      <Grid container spacing={2}>
        <Grid item xs={12}>
          <Card>
            <FlatList
              headerNote={`Showing times for schedule in ${timeZone}.`}
              emptyMessage='No notification rules.'
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
              items={q.data.schedule.onCallNotificationRules.map((rule) => {
                const display = rule.dest.display
                return {
                  icon: (
                    <DestinationAvatar
                      iconURL={display.iconURL}
                      iconAltText={display.iconAltText}
                    />
                  ),
                  title: display.text,
                  subText: 'Notifies ' + onCallRuleSummary(timeZone, rule),
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
      <Suspense>
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
      </Suspense>
    </React.Fragment>
  )
}
