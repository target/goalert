import React, { useState } from 'react'
import { PropTypes as p } from 'prop-types'
import gql from 'graphql-tag'
import { useQuery, useMutation } from '@apollo/react-hooks'
import FormDialog from '../../dialogs/FormDialog'
import CalendarSubscribeForm, {
  reminderMinutesOptions,
} from './CalendarSubscribeForm'
import { GenericError, ObjectNotFound } from '../../error-pages'
import _ from 'lodash-es'
import Spinner from '../../loading/components/Spinner'
import { fieldErrors } from '../../util/errutil'

const query = gql`
  query($id: ID!) {
    userCalendarSubscription(id: $id) {
      id
      name
      reminderMinutes
      scheduleID
    }
  }
`

const mutation = gql`
  mutation($input: UpdateUserCalendarSubscriptionInput!) {
    updateUserCalendarSubscription(input: $input)
  }
`

/*
 * Load edit data here before rendering edit content to
 * avoid breaking any rules of hooks
 */
export default function CalendarSubscribeEditDialog(props) {
  const { data, loading, error } = useQuery(query, {
    variables: { id: props.calSubscriptionID },
  })

  if (error) return <GenericError error={error.message} />
  if (!_.get(data, 'userCalendarSubscription.id')) {
    return loading ? <Spinner /> : <ObjectNotFound />
  }

  return (
    <CalendarSubscribeEditDialogContent
      data={data.userCalendarSubscription}
      onClose={props.onClose}
    />
  )
}

CalendarSubscribeEditDialog.propTypes = {
  calSubscriptionID: p.string.isRequired,
  onClose: p.func.isRequired,
}

export function CalendarSubscribeEditDialogContent(props) {
  const { data, onClose } = props

  // set default values from retrieved data
  const [value, setValue] = useState({
    name: _.get(data, 'name', ''),
    scheduleID: _.get(data, 'scheduleID', null),
    reminderMinutes: _.get(data, 'reminderMinutes', []).map(r =>
      reminderMinutesOptions.find(opt => opt.value === r),
    ),
  })

  // setup the mutation
  const [updateSubscription, updateSubscriptionStatus] = useMutation(mutation, {
    variables: {
      input: {
        id: props.data.id,
        name: value.name,
        reminderMinutes: value.reminderMinutes.map(r => r && r.value),
      },
    },
    onCompleted: () => props.onClose(),
  })

  return (
    <FormDialog
      title='Edit Calendar Subscription'
      onClose={onClose}
      loading={updateSubscriptionStatus.loading}
      onSubmit={() => updateSubscription()}
      form={(
        <CalendarSubscribeForm
          errors={fieldErrors(updateSubscriptionStatus.error)}
          loading={updateSubscriptionStatus.loading}
          onChange={setValue}
          value={value}
          scheduleReadOnly
        />
      )}
    />
  )
}

CalendarSubscribeEditDialogContent.propTypes = {
  data: p.object.isRequired,
  onClose: p.func.isRequired,
}
