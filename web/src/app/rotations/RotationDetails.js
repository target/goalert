import React from 'react'
import p from 'prop-types'
import gql from 'graphql-tag'
import PageActions from '../util/PageActions'
import Query from '../util/Query'
import OtherActions from '../util/OtherActions'
import CreateFAB from '../lists/CreateFAB'
import { handoffSummary } from './util'
import DetailsPage from '../details/DetailsPage'
import RotationEditDialog from './RotationEditDialog'
import RotationDeleteDialog from './RotationDeleteDialog'
import RotationUserList from './RotationUserList'
import RotationAddUserDialog from './RotationAddUserDialog'
import { QuerySetFavoriteButton } from '../util/QuerySetFavoriteButton'

const query = gql`
  query rotationDetails($rotationID: ID!) {
    rotation(id: $rotationID) {
      id
      name
      description
      activeUserIndex
      userIDs
      type
      shiftLength
      timeZone
      start
    }
  }
`

const partialQuery = gql`
  query($rotationID: ID!) {
    rotation(id: $rotationID) {
      id
      name
      description
    }
  }
`

export default class RotationDetails extends React.PureComponent {
  static propTypes = {
    rotationID: p.string.isRequired,
  }

  state = {
    value: null,
    edit: false,
    delete: false,
    addUser: false,
  }

  render() {
    return (
      <Query
        query={query}
        partialQuery={partialQuery}
        variables={{ rotationID: this.props.rotationID }}
        render={this.renderData}
      />
    )
  }

  renderData = ({ data }) => {
    const summary = handoffSummary(data.rotation)
    return (
      <React.Fragment>
        <PageActions>
          <QuerySetFavoriteButton rotationID={data.rotation.id} />
          <OtherActions
            actions={[
              {
                label: 'Edit Rotation',
                onClick: () => this.setState({ edit: true }),
              },
              {
                label: 'Delete Rotation',
                onClick: () => this.setState({ delete: true }),
              },
            ]}
          />
        </PageActions>
        <DetailsPage
          title={data.rotation.name}
          details={data.rotation.description}
          titleFooter={summary}
          pageFooter={<RotationUserList rotationID={this.props.rotationID} />}
        />

        <CreateFAB onClick={() => this.setState({ addUser: true })} />
        {this.state.addUser && (
          <RotationAddUserDialog
            rotationID={this.props.rotationID}
            userIDs={data.rotation.userIDs}
            onClose={() => this.setState({ addUser: false })}
          />
        )}
        {this.state.edit && (
          <RotationEditDialog
            onClose={() => this.setState({ edit: false })}
            rotationID={this.props.rotationID}
          />
        )}
        {this.state.delete && (
          <RotationDeleteDialog
            onClose={() => this.setState({ delete: false })}
            rotationID={this.props.rotationID}
          />
        )}
      </React.Fragment>
    )
  }
}
