import React from 'react'
import p from 'prop-types'
import { graphql2Client } from '../apollo'
import gql from 'graphql-tag'
import { Mutation } from 'react-apollo'
import { nonFieldErrors } from '../util/errutil'
import { Redirect } from 'react-router'
import Query from '../util/Query'
import FormDialog from '../dialogs/FormDialog'

const query = gql`
  query($id: ID!) {
    rotation(id: $id) {
      id
      name
      description
      start
      timeZone
      type
    }
  }
`
const mutation = gql`
  mutation($input: [TargetInput!]!) {
    deleteAll(input: $input)
  }
`

export default class RotationDeleteDialog extends React.PureComponent {
  static propTypes = {
    rotationID: p.string.isRequired,
    onClose: p.func,
  }

  render() {
    return (
      <Query
        noPoll
        query={query}
        variables={{ id: this.props.rotationID }}
        render={({ data, error }) => this.renderMutation(data.rotation)}
      />
    )
  }

  renderMutation(rotData) {
    return (
      <Mutation client={graphql2Client} mutation={mutation}>
        {(commit, status) => this.renderDialog(rotData, commit, status)}
      </Mutation>
    )
  }

  renderDialog(rotData, commit, mutStatus) {
    const { loading, error, data } = mutStatus
    if (data && data.deleteAll) {
      return <Redirect push to={`/rotations`} />
    }
    return (
      <FormDialog
        title='Are you sure?'
        confirm
        subTitle={`This will delete the rotation: ${rotData.name}`}
        loading={loading}
        errors={nonFieldErrors(error)}
        onClose={this.props.onClose}
        onSubmit={() => {
          const input = [
            {
              type: 'rotation',
              id: this.props.rotationID,
            },
          ]
          return commit({
            variables: {
              input,
            },
          })
        }}
      />
    )
  }
}
