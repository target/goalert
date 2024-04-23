import React from 'react'
import { gql, useQuery, useMutation } from 'urql'
import FormDialog from '../dialogs/FormDialog'
import { startCase } from 'lodash'
import { GenericError } from '../error-pages'
import Spinner from '../loading/components/Spinner'

interface ScheduleRuleDeleteDialogProps {
  scheduleID: string
  target: Target
  onClose: () => void
}

interface Target {
  id: string
  type: string
}

const query = gql`
  query ($id: ID!, $tgt: TargetInput!) {
    schedule(id: $id) {
      id
      target(input: $tgt) {
        target {
          id
          name
          type
        }
      }
    }
  }
`

const mutation = gql`
  mutation ($input: ScheduleTargetInput!) {
    updateScheduleTarget(input: $input)
  }
`

export default function ScheduleRuleDeleteDialog(
  props: ScheduleRuleDeleteDialogProps,
): JSX.Element {
  const [{ data, error, fetching }] = useQuery({
    query,
    variables: { id: props.scheduleID, tgt: props.target },
  })

  const [deleteRuleStatus, deleteRule] = useMutation(mutation)

  if (error) {
    return <GenericError error={error.message} />
  }

  if (fetching && !data) {
    return <Spinner />
  }

  return (
    <FormDialog
      onClose={props.onClose}
      title={`Remove ${startCase(props.target.type)} From Schedule?`}
      subTitle={`This will remove all rules, as well as end any active or future on-call shifts on this schedule for ${props.target.type}: ${data.schedule.target.target.name}.`}
      caption='Overrides will not be affected.'
      loading={deleteRuleStatus.fetching}
      errors={deleteRuleStatus.error ? [deleteRuleStatus.error] : []}
      confirm
      onSubmit={() => {
        deleteRule(
          {
            input: {
              target: props.target,
              scheduleID: props.scheduleID,
              rules: [],
            },
          },
          { additionalTypenames: ['ScheduleTargetInput', 'Schedule'] },
        ).then((result) => {
          if (!result.error) props.onClose()
        })
      }}
    />
  )
}
