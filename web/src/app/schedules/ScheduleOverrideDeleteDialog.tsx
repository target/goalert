import React from 'react'

import { gql, useQuery, useMutation } from 'urql'
import { nonFieldErrors } from '../util/errutil'
import { Typography } from '@mui/material'
import FormDialog from '../dialogs/FormDialog'
import { useURLParam } from '../actions/hooks'
import { formatOverrideTime } from './util'
import { GenericError } from '../error-pages'
import Spinner from '../loading/components/Spinner'

const query = gql`
  query ($id: ID!) {
    userOverride(id: $id) {
      id
      start
      end
      addUser {
        id
        name
      }
      removeUser {
        id
        name
      }
    }
  }
`

const mutation = gql`
  mutation delete($input: [TargetInput!]!) {
    deleteAll(input: $input)
  }
`

export default function ScheduleOverrideDeleteDialog(props: {
  overrideID: string
  onClose: () => void
}): React.ReactNode {
  const [zone] = useURLParam('tz', 'local')

  const [{ data, fetching, error }] = useQuery({
    query,
    variables: { id: props.overrideID },
  })

  const [deleteOverrideStatus, deleteOverride] = useMutation(mutation)

  if (error) {
    return <GenericError error={error.message} />
  }

  if (fetching && !data) {
    return <Spinner />
  }

  const addUser = data.userOverride.addUser ? data.userOverride.addUser : ''
  const removeUser = data.userOverride.removeUser
    ? data.userOverride.removeUser
    : ''
  const start = data.userOverride.start ? data.userOverride.start : ''
  const end = data.userOverride.end ? data.userOverride.end : ''

  const isReplace = addUser && removeUser
  const verb = addUser ? 'Added' : 'Removed'
  const time = formatOverrideTime(start, end, zone)

  const caption = isReplace
    ? `Replaced ${removeUser.name} from ${time}`
    : `${verb} from ${time}`

  return (
    <FormDialog
      title='Are you sure?'
      confirm
      subTitle={`This will delete the override for: ${
        addUser ? addUser.name : removeUser.name
      }`}
      loading={deleteOverrideStatus.fetching}
      errors={nonFieldErrors(deleteOverrideStatus.error)}
      onClose={props.onClose}
      onSubmit={() =>
        deleteOverride(
          {
            input: [
              {
                type: 'userOverride',
                id: props.overrideID,
              },
            ],
          },
          { additionalTypenames: ['UserOverride'] },
        ).then((res) => {
          if (res.error) return
          props.onClose()
        })
      }
      form={<Typography variant='caption'>{caption}</Typography>}
    />
  )
}
