import React from 'react'
import p from 'prop-types'

import { gql } from '@apollo/client'
import { Mutation } from '@apollo/client/react/components'
import { nonFieldErrors } from '../util/errutil'
import Query from '../util/Query'
import { Typography } from '@material-ui/core'
import FormDialog from '../dialogs/FormDialog'
import { useURLParam } from '../actions/hooks'
import { formatOverrideTime } from './util'

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
  mutation ($id: ID!) {
    deleteAll(input: [{ type: userOverride, id: $id }])
  }
`

export default function ScheduleOverrideDeleteDialog({ overrideID, onClose }) {
  const [zone] = useURLParam('tz', 'local')

  function renderDialog(data, commit, mutStatus) {
    const { loading, error } = mutStatus
    const { addUser, removeUser, start, end } = data

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
        loading={loading}
        errors={nonFieldErrors(error)}
        onClose={onClose}
        onSubmit={() => {
          return commit({
            variables: {
              id: overrideID,
            },
          })
        }}
        form={<Typography variant='caption'>{caption}</Typography>}
      />
    )
  }

  function renderMutation(data) {
    return (
      <Mutation mutation={mutation} onCompleted={onClose}>
        {(commit, status) => renderDialog(data, commit, status)}
      </Mutation>
    )
  }

  function renderQuery() {
    return (
      <Query
        noPoll
        query={query}
        variables={{ id: overrideID }}
        render={({ data }) => renderMutation(data.userOverride)}
      />
    )
  }
  return renderQuery()
}

ScheduleOverrideDeleteDialog.propTypes = {
  overrideID: p.string.isRequired,
  onClose: p.func,
}
