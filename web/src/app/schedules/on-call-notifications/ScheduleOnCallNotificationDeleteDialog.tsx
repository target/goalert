import React, { useContext } from 'react'
import { useMutation } from '@apollo/client'
import FormDialog from '../../dialogs/FormDialog'
import { nonFieldErrors } from '../../util/errutil'
import { ScheduleContext, setMutation } from './ScheduleOnCallNotificationsList'
import { Rule, mapDataToInput, getDayNames } from './util'
import { useURLParam } from '../../actions/hooks'
import { DateTime } from 'luxon'

function getDeleteSummary(
  r: Rule,
  scheduleZone: string,
  displayZone: string,
): string {
  const prefix = `${r.target.name} will no longer be notified`

  if (r.time && r.weekdayFilter) {
    const timeStr = DateTime.fromFormat(r.time, 'HH:mm', {
      zone: scheduleZone,
    })
      .setZone(displayZone)
      .toFormat('h:mm a ZZZZ')

    return `${prefix} ${getDayNames(r.weekdayFilter)} at ${timeStr}`
  }

  return `${prefix} when on-call changes.`
}

export function getRuleSummary(
  rule: Rule,
  scheduleZone: string,
  displayZone: string,
): string {
  if (rule.time && rule.weekdayFilter) {
    const timeStr = DateTime.fromFormat(rule.time, 'HH:mm', {
      zone: scheduleZone,
    })
      .setZone(displayZone)
      .toFormat('h:mm a ZZZZ')

    return `Notifies ${getDayNames(rule.weekdayFilter)} at ${timeStr}`
  }

  return 'Notifies when on-call hands off'
}

interface ScheduleOnCallNotificationDeleteDialogProps {
  rule: Rule
  onClose: () => void
}

export default function ScheduleOnCallNotificationDeleteDialog(
  p: ScheduleOnCallNotificationDeleteDialogProps,
): JSX.Element {
  const [URLZone] = useURLParam('tz', 'local')
  const schedCtx = useContext(ScheduleContext)

  const [mutate, mutationStatus] = useMutation(setMutation, {
    variables: {
      input: {
        scheduleID: schedCtx.id,
        rules: mapDataToInput(
          schedCtx.onCallNotificationRules.filter(
            (nr: Rule) => nr.id !== p.rule.id,
          ),
          schedCtx.timeZone,
        ),
      },
    },
    onCompleted: () => p.onClose(),
  })

  return (
    <FormDialog
      title='Are you sure?'
      confirm
      loading={mutationStatus.loading}
      errors={nonFieldErrors(mutationStatus.error)}
      subTitle={getDeleteSummary(p.rule, schedCtx.timeZone, URLZone)}
      onSubmit={() => mutate()}
      onClose={() => p.onClose()}
    />
  )
}
