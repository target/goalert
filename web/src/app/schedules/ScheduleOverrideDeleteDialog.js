import React from 'react'
import p from 'prop-types'

import { connect } from 'react-redux'
import { gql } from '@apollo/client'
import { Mutation } from '@apollo/client/react/components'
import { nonFieldErrors } from '../util/errutil'
import Query from '../util/Query'
import { Typography } from '@material-ui/core'
import FormDialog from '../dialogs/FormDialog'
import { urlParamSelector } from '../selectors'
import { formatOverrideTime } from './util'

const query = gql`
  query($id: ID!) {
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
  mutation($id: ID!) {
    deleteAll(input: [{ type: userOverride, id: $id }])
  }
`

@connect((state) => ({ zone: urlParamSelector(state)('tz') }))
export default class ScheduleOverrideDeleteDialog extends React.PureComponent {
  static propTypes = {
    overrideID: p.string.isRequired,
    onClose: p.func,
  }

  renderQuery() {
    return (
      <Query
        noPoll
        query={query}
        variables={{ id: this.props.overrideID }}
        render={({ data }) => this.renderMutation(data.userOverride)}
      />
    )
  }

  renderMutation(data) {
    return (
      <Mutation mutation={mutation} onCompleted={this.props.onClose}>
        {(commit, status) => this.renderDialog(data, commit, status)}
      </Mutation>
    )
  }

  renderDialog(data, commit, mutStatus) {
    const { loading, error } = mutStatus

    const zone = this.props.zone
    const isReplace = data.addUser && data.removeUser
    const verb = data.addUser ? 'Added' : 'Removed'

    const time = formatOverrideTime(data.start, data.end, zone)

    const caption = isReplace
      ? `Replaced ${data.removeUser.name} from ${time}`
      : `${verb} from ${time}`
    return (
      <FormDialog
        title='Are you sure?'
        confirm
        subTitle={`This will delete the override for: ${
          data.addUser ? data.addUser.name : data.removeUser.name
        }`}
        loading={loading}
        errors={nonFieldErrors(error)}
        onClose={this.props.onClose}
        onSubmit={() => {
          return commit({
            variables: {
              id: this.props.overrideID,
            },
          })
        }}
        form={<Typography variant='caption'>{caption}</Typography>}
      />
    )
  }

  render() {
    return this.renderQuery()
  }
}
