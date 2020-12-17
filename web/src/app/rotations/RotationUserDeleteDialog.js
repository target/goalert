import React from 'react'
import { gql } from '@apollo/client'
import p from 'prop-types'
import Query from '../util/Query'
import { Mutation } from '@apollo/client/react/components'
import FormDialog from '../dialogs/FormDialog'

const query = gql`
  query($id: ID!) {
    rotation(id: $id) {
      id
      userIDs
      users {
        id
        name
      }
      activeUserIndex
    }
  }
`

const mutation = gql`
  mutation($input: UpdateRotationInput!) {
    updateRotation(input: $input)
  }
`
export default class RotationUserDeleteDialog extends React.PureComponent {
  static propTypes = {
    rotationID: p.string.isRequired,
    userIndex: p.number.isRequired,
    onClose: p.func.isRequired,
  }

  render() {
    return (
      <Query
        query={query}
        variables={{ id: this.props.rotationID }}
        render={({ data }) => this.renderMutation(data.rotation)}
      />
    )
  }

  renderMutation(data) {
    return (
      <Mutation mutation={mutation} onCompleted={this.props.onClose}>
        {(commit) => this.renderDialog(data, commit)}
      </Mutation>
    )
  }

  renderDialog(data, commit) {
    const { userIDs, users } = data
    const { rotationID, userIndex, onClose } = this.props

    return (
      <FormDialog
        title='Are you sure?'
        confirm
        subTitle={`This will delete ${
          users[userIndex] ? users[userIndex].name : null
        } from this rotation.`}
        onClose={onClose}
        onSubmit={() => {
          return commit({
            variables: {
              input: {
                id: rotationID,
                userIDs: userIDs.filter((id, index) => index !== userIndex),
              },
            },
          })
        }}
      />
    )
  }
}
