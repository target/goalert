import React, { Suspense, useState } from 'react'
import { Button, Grid, Card, Typography, Tooltip } from '@mui/material'
import FlatList from '../../lists/FlatList'
import OtherActions from '../../util/OtherActions'
import { onCallRuleSummary } from './util'
import ScheduleOnCallNotificationsCreateDialog from './ScheduleOnCallNotificationsCreateDialog'
import ScheduleOnCallNotificationsDeleteDialog from './ScheduleOnCallNotificationsDeleteDialog'
import CreateFAB from '../../lists/CreateFAB'
import ScheduleOnCallNotificationsEditDialog from './ScheduleOnCallNotificationsEditDialog'
import { useIsWidthDown } from '../../util/useWidth'
import { Add } from '@mui/icons-material'
import Error from '@mui/icons-material/Error'
import { gql, useQuery } from 'urql'
import { Schedule } from '../../../schema'
import { DestinationAvatar } from '../../util/DestinationAvatar'
import { styles as globalStyles } from '../../styles/materialStyles'
import makeStyles from '@mui/styles/makeStyles'

export type ScheduleOnCallNotificationsListDestProps = {
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
            text
            iconURL
            iconAltText
          }
        }
      }
    }
  }
`

const useStyles = makeStyles((theme) => ({
  ...globalStyles(theme),
}))

export default function ScheduleOnCallNotificationsListDest({
  scheduleID,
}: ScheduleOnCallNotificationsListDestProps): JSX.Element {
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
            <FlatList
              headerNote={
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
                const display = rule.dest.displayInfo
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
