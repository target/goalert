import React, { useState } from 'react'
import { PropTypes as p } from 'prop-types'
import gql from 'graphql-tag'
import { useMutation } from '@apollo/react-hooks'
import FormDialog from '../../dialogs/FormDialog'
import { getForm, FormTitle, getSubtitle } from './formHelper'
import CalendarSubscribeForm from './CalendarSubscribeForm'

const SUBTITLE =
  'Editing the schedule or alarm will result in a new URL being generated.'

// todo: update input and names
const query = gql`
  query($id: ID!) {
    schedule(id: $id) {
      id
      name
      description
      timeZone
    }
  }
`

const mutation = gql`
  mutation($input: UpdateScheduleInput!) {
    updateSchedule(input: $input)
  }
`

export default function CalendarSubscribeEditDialog(props) {
  const [value, setValue] = useState(null)
  const [isComplete, setIsComplete] = useState(false)

  const queryVariables = { id: props.calSubscriptionID }
  // const { data, loading, error } = useQuery(query, {
  //   variables: queryVariables,
  // })

  const [updateSubscription] = useMutation(mutation, {
    variables: {
      input: {
        id: props.calSubscriptionID,
        ...value,
      },
    },
    onCompleted: () => {
      // todo: if only name was updated, skip showing success form
      setIsComplete(true)
      props.onClose()
    },
    refetchQueries: () => [
      {
        query,
        variables: queryVariables,
      },
    ],
  })

  const form = <CalendarSubscribeForm onChange={setValue} value={value} />

  return (
    <FormDialog
      title={FormTitle(isComplete, 'Create New Calendar Subscription')}
      subTitle={getSubtitle(isComplete, SUBTITLE)}
      onClose={props.onClose}
      alert={isComplete}
      primaryActionLabel={isComplete ? 'Done' : null}
      onSubmit={() => (isComplete ? props.onClose() : updateSubscription())}
      form={getForm(isComplete, form, 'url')}
    />
  )
}

CalendarSubscribeEditDialog.propTypes = {
  calSubscriptionID: p.string.isRequired,
  onClose: p.func.isRequired,
}
