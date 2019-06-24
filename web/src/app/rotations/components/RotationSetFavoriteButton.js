import React from 'react'
import p from 'prop-types'
import gql from 'graphql-tag'
import { graphql2Client } from '../../apollo'
import Query from '../../util/Query'
import { Mutation } from 'react-apollo'
import SetFavoriteButton from '../../util/SetFavoriteButton'

const query = gql`
  query favQuery($id: ID!) {
    rotation(id: $id) {
      id
      isFavorite
    }
  }
`

const mutation = gql`
  mutation setFav($input: SetFavoriteInput!) {
    setFavorite(input: $input)
  }
`

export default class RotationSetFavoriteButton extends React.PureComponent {
  static propTypes = {
    rotationID: p.string.isRequired,
  }
  render() {
    return (
      <Query
        query={query}
        variables={{ id: this.props.rotationID }}
        render={({ data }) => {
          if (!data || !data.rotation) return null

          return this.renderMutation(
            data && data.rotation && data.rotation.isFavorite,
          )
        }}
      />
    )
  }

  renderMutation(isFavorite) {
    return (
      <Mutation
        key='main'
        mutation={mutation}
        client={graphql2Client}
        awaitRefetchQueries
        refetchQueries={['favQuery']}
      >
        {mutation => this.renderSetFavButton(isFavorite, mutation)}
      </Mutation>
    )
  }

  renderSetFavButton(isFavorite, mutation) {
    const { rotationID: id } = this.props
    return (
      <SetFavoriteButton
        type={'rotation'}
        isFavorite={isFavorite}
        onSubmit={() => {
          return mutation({
            variables: {
              input: {
                target: { id, type: 'rotation' },
                favorite: !isFavorite,
              },
            },
          })
        }}
      />
    )
  }
}
