import React from 'react'
import p from 'prop-types'
import gql from 'graphql-tag'
import { graphql2Client } from '../../apollo'
import Query from '../../util/Query'
import { Mutation } from 'react-apollo'
import { SetFavoriteButton } from '../../util/SetFavoriteButton'

const query = gql`
  query favQuery($id: ID!) {
    service(id: $id) {
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

export function ServicesSetFavoriteButton(props) {
  return (
    <Query
      query={query}
      variables={{ id: props.serviceID }}
      render={({ data }) => {
        if (!data || !data.service) return null

        return renderMutation(data.service.isFavorite, props.serviceID)
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
      type={'service'}
      isFavorite={isFavorite}
      onSubmit={() => {
        return mutation({
          variables: {
            input: {
              target: { id, type: 'service' },
              favorite: !isFavorite,
            },
          },
        })
      }}
    />
  )
}

ServicesSetFavoriteButton.propTypes = {
  serviceID: p.string.isRequired,
}
