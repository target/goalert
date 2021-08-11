import React, { useState } from 'react'
import { gql, useMutation } from '@apollo/client'
import p from 'prop-types'
import Button from '@material-ui/core/Button'
import FormDialog from '../dialogs/FormDialog'
import { DateTime } from 'luxon'
import ScheduleOverrideForm from './ScheduleOverrideForm'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import useOverrideNotices from './useOverrideNotices'

const copyText = {
  add: {
    title: 'Temporarily Add a User',
    desc: 'This will add a new shift for the selected user, while the override is active. Existing shifts will remain unaffected.',
  },
  remove: {
    title: 'Temporarily Remove a User',
    desc: 'This will remove (or split/shorten) shifts belonging to the selected user, while the override is active.',
  },
  replace: {
    title: 'Temporarily Replace a User',
    desc: 'This will replace the selected user with another during any existing shifts, while the override is active. No new shifts will be created, only who is on-call will be changed.',
  },
  choose: {
    title: 'Choose an override action',
    desc: 'bla bla bla desc',
  },
}

const mutation = gql`
  mutation ($input: CreateUserOverrideInput!) {
    createUserOverride(input: $input) {
      id
    }
  }
`

export default function ScheduleOverrideCreateDialog(props) {
  const [value, setValue] = useState({
    addUserID: '',
    removeUserID: '',
    start: DateTime.local().startOf('hour').toISO(),
    end: DateTime.local().startOf('hour').plus({ hours: 8 }).toISO(),
    ...props.defaultValue,
  })

  const notices = useOverrideNotices(props.scheduleID, value)

  const [mutate, { loading, error }] = useMutation(mutation, {
    variables: {
      input: {
        ...value,
        scheduleID: props.scheduleID,
      },
    },
    onCompleted: props.onClose,
  })

  function handleChoose(type) {
    console.log('SPENCER', type)
    props.onChooseOverrideType({
      variant: type,
      // defaultValue: {
      //   addUserID: '',
      //   removeUserID: '',
      //   start: DateTime.local().startOf('hour').toISO(),
      //   end: DateTime.local().startOf('hour').plus({ hours: 8 }).toISO(),
      //   ...props.defaultValue,
      // },
    })
  }

  function renderFormContent() {
    if (props.variant === 'choose') {
      return (
        <div>
          <Button color='primary' onClick={() => handleChoose('add')}>
            Add person to shift
          </Button>
          <Button color='primary' onClick={handleChoose('remove')}>
            Remove person from shift
          </Button>
          <Button color='primary' onClick={handleChoose('replace')}>
            Replace person on shift
          </Button>
        </div>
      )
    }

    return (
      <ScheduleOverrideForm
        add={value.variant !== 'remove'}
        remove={value.variant !== 'add'}
        scheduleID={props.scheduleID}
        disabled={loading}
        errors={fieldErrors(error)}
        value={value}
        onChange={(newValue) => setValue(newValue)}
        removeUserReadOnly={props.removeUserReadOnly}
      />
    )
  }

  return (
    <FormDialog
      onClose={props.onClose}
      title={copyText[props.variant].title}
      subTitle={copyText[props.variant].desc}
      errors={nonFieldErrors(error)}
      notices={notices} // create and edit dialogue
      onSubmit={() => mutate()}
      form={renderFormContent()}
    />
  )
}

ScheduleOverrideCreateDialog.defaultProps = {
  defaultValue: {},
}

ScheduleOverrideCreateDialog.propTypes = {
  scheduleID: p.string.isRequired,
  variant: p.oneOf(['add', 'remove', 'replace', 'choose']).isRequired,
  onClose: p.func,
  removeUserReadOnly: p.bool,
  defaultValue: p.shape({
    addUserID: p.string,
    removeUserID: p.string,
    start: p.string,
    end: p.string,
  }),
  onChooseOverrideType: p.func,
}
