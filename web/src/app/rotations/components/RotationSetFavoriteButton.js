import React from 'react'
import p from 'prop-types'
import gql from 'graphql-tag'
import { graphql2Client } from '../../apollo'
import Query from '../../util/Query'
import { Mutation } from 'react-apollo'
import { SetFavoriteButton } from '../../util/SetFavoriteButton'

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

export function RotationSetFavoriteButton(props) {
  return (
    <Query
      query={query}
      variables={{ id: props.rotationID }}
      render={({ data }) => {
        if (!data || !data.rotation) return null

        return renderMutation(data.rotation.isFavorite, props.rotationID)
      }}
    />
  )
}

function renderMutation(isFavorite, id) {
  return (
    <Mutation
      mutation={mutation}
      client={graphql2Client}
      awaitRefetchQueries
      refetchQueries={['favQuery']}
    >
      {mutation => renderSetFavButton(isFavorite, mutation, id)}
    </Mutation>
  )
}

function renderSetFavButton(isFavorite, mutation, id) {
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

RotationSetFavoriteButton.propTypes = {
  rotationID: p.string.isRequired,
}
