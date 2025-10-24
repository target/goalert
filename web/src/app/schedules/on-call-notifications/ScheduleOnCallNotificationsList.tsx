import React, { Suspense, useState } from 'react'
import { Button, Grid, Card, Typography, Tooltip, Theme } from '@mui/material'
import OtherActions from '../../util/OtherActions'
import { onCallRuleSummary } from './util'
import ScheduleOnCallNotificationsDeleteDialog from './ScheduleOnCallNotificationsDeleteDialog'
import CreateFAB from '../../lists/CreateFAB'
import { useIsWidthDown } from '../../util/useWidth'
import { Add } from '@mui/icons-material'
import Error from '@mui/icons-material/Error'
import { gql, useQuery } from 'urql'
import { Schedule } from '../../../schema'
import { DestinationAvatar } from '../../util/DestinationAvatar'
import { styles as globalStyles } from '../../styles/materialStyles'
import makeStyles from '@mui/styles/makeStyles'
import ScheduleOnCallNotificationsCreateDialog from './ScheduleOnCallNotificationsCreateDialog'
import ScheduleOnCallNotificationsEditDialog from './ScheduleOnCallNotificationsEditDialog'
import CompList from '../../lists/CompList'
import { CompListItemText } from '../../lists/CompListItems'

export type ScheduleOnCallNotificationsListProps = {
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
          displayInfo {
            ... on DestinationDisplayInfo {
              text
              iconURL
              iconAltText
            }
            ... on DestinationDisplayInfoError {
              error
            }
          }
        }
      }
    }
  }
`

const useStyles = makeStyles((theme: Theme) => ({
  ...globalStyles(theme),
}))

export default function ScheduleOnCallNotificationsList({
  scheduleID,
}: ScheduleOnCallNotificationsListProps): JSX.Element {
  const [createRule, setCreateRule] = useState(false)
  const [editRuleID, setEditRuleID] = useState('')
  const [deleteRuleID, setDeleteRuleID] = useState('')
  const isMobile = useIsWidthDown('md')

  const classes = useStyles()

  const [q] = useQuery<{ schedule: Schedule }>({
    query,
    variables: { scheduleID },
  })

  if (q.error || !q.data) {
    return (
      <Grid container spacing={2}>
        <Grid item xs={12}>
          <Card>
            <CompList
              note={
                <Typography
                  component='p'
                  variant='subtitle1'
                  style={{ display: 'flex' }}
                >
                  <Tooltip title={q.error?.message}>
                    <Error className={classes.error} />
                  </Tooltip>
                  &nbsp;
                  <span className={classes.error}>
                    Error loading notification rules.
                  </span>
                </Typography>
              }
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
            <CompList
              note={`Showing times for schedule in ${timeZone}.`}
              emptyMessage='No notification rules.'
              hideActionOnMobile
              action={
                <Button
                  variant='contained'
                  onClick={() => setCreateRule(true)}
                  startIcon={<Add />}
                >
                  Create Notification Rule
                </Button>
              }
            >
              {q.data.schedule.onCallNotificationRules.map((rule) => {
                const display = rule.dest.displayInfo
                if ('error' in display) {
                  return (
                    <CompListItemText
                      key={rule.id}
                      icon={<DestinationAvatar error />}
                      title={`ERROR: ${display.error}`}
                      subText={`Notifies ${onCallRuleSummary(timeZone, rule)}`}
                      action={
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
                      }
                    />
                  )
                }

                return (
                  <CompListItemText
                    key={rule.id}
                    icon={
                      <DestinationAvatar
                        iconURL={display.iconURL}
                        iconAltText={display.iconAltText}
                      />
                    }
                    title={display.text}
                    subText={'Notifies ' + onCallRuleSummary(timeZone, rule)}
                    action={
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
                    }
                  />
                )
              })}
            </CompList>
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
