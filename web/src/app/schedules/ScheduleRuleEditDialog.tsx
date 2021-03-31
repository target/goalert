import React, { useState } from 'react'
import { gql, useMutation, useQuery } from '@apollo/client'
import FormDialog from '../dialogs/FormDialog'
import ScheduleRuleForm from './ScheduleRuleForm'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import _ from 'lodash'
import { gqlClockTimeToISO, isoToGQLClockTime } from './util'
import { Mutation, Query, ScheduleTargetInput, Target } from '../../schema'
import Spinner from '../loading/components/Spinner'

const query = gql`
  query($id: ID!, $tgt: TargetInput!) {
    schedule(id: $id) {
      id
      timeZone
      target(input: $tgt) {
        rules {
          id
          start
          end
          weekdayFilter
        }
      }
    }
  }
`

const mutation = gql`
  mutation($input: ScheduleTargetInput!) {
    updateScheduleTarget(input: $input)
  }
`

interface ScheduleRuleEditDialogProps {
  scheduleID: string
  target: Target
  onClose: () => void
}

function ScheduleRuleEditDialog(
  props: ScheduleRuleEditDialogProps,
): JSX.Element {
  const [value, setValue] = useState<ScheduleTargetInput | null>(null)

  const { data: qData, loading: qLoading, error: qError } = useQuery<Query>(
    query,
    {
      variables: { id: props.scheduleID, tgt: props.target },
      pollInterval: 0,
      fetchPolicy: 'network-only',
    },
  )

  // shouldComponentUpdate(nextProps, nextState) {
  //   if (this.state !== nextState) return true

  //   return false
  // }

  const [mutate, { loading: mLoading, error: mError }] = useMutation<Mutation>(
    mutation,
    {
      onCompleted: () => props.onClose(),
    },
  )

  if (qLoading && !qData) {
    return <Spinner />
  }

  const defaultValue: ScheduleTargetInput = {
    target: props.target,
    scheduleID: props.scheduleID,
    rules:
      qData?.schedule?.target?.rules.map((r) => ({
        id: r.id,
        weekdayFilter: r.weekdayFilter,
        start: gqlClockTimeToISO(r.start, qData?.schedule?.timeZone),
        end: gqlClockTimeToISO(r.end, qData?.schedule?.timeZone),
      })) ?? [],
  }

  return (
    <FormDialog
      onClose={props.onClose}
      title={`Edit Rules for ${_.startCase(props.target.type)}`}
      loading={mLoading}
      errors={nonFieldErrors(qError || mError)}
      maxWidth='md'
      onSubmit={() => {
        if (!value) {
          // no changes
          props.onClose()
          return
        }
        mutate({
          variables: {
            input: {
              target: props.target,
              scheduleID: props.scheduleID,
              rules: value.rules.map((r) => ({
                ...r,
                start: isoToGQLClockTime(r.start, qData?.schedule?.timeZone),
                end: isoToGQLClockTime(r.end, qData?.schedule?.timeZone),
              })),
            },
          },
        })
      }}
      form={
        <ScheduleRuleForm
          targetType={props.target.type}
          targetDisabled // since we are in edit mode
          scheduleID={props.scheduleID}
          disabled={qLoading || mLoading}
          errors={fieldErrors(qError || mError)}
          value={value || defaultValue}
          onChange={(value: ScheduleTargetInput) => setValue(value)}
        />
      }
    />
  )
}

export default ScheduleRuleEditDialog
