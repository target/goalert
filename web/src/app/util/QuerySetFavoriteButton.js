import React from 'react'
import p from 'prop-types'
import gql from 'graphql-tag'
import { graphql2Client } from '../apollo'
import Query from '../util/Query'
import { Mutation } from 'react-apollo'
import { SetFavoriteButton } from './SetFavoriteButton'
import { oneOfShape } from '../util/propTypes'

const queries = {
  service: gql`
    query serviceFavQuery($id: ID!) {
      service(id: $id) {
        id
        isFavorite
      }
    }
  `,
  rotation: gql`
    query rotationFavQuery($id: ID!) {
      rotation(id: $id) {
        id
        isFavorite
      }
    }
  `,
}

const mutation = gql`
  mutation setFav($input: SetFavoriteInput!) {
    setFavorite(input: $input)
  }
`

export function QuerySetFavoriteButton(props) {
  let typeName = ''
  let id = ''
  if (props.rotationID) {
    typeName = 'rotation'
    id = props.rotationID
  } else if (props.serviceID) {
    typeName = 'service'
    id = props.serviceID
  } else {
    return null
  }
  return (
    <Query
      query={queries[typeName]}
      variables={{ id: id }}
      render={({ data }) => {
        if (!data || !data[typeName]) return null

        return renderMutation(data[typeName].isFavorite, id, typeName)
      }}
    />
  )
}

function renderMutation(isFavorite, id, typeName) {
  return (
    <Mutation
      mutation={mutation}
      client={graphql2Client}
      awaitRefetchQueries
      refetchQueries={[`${typeName}FavQuery`]}
    >
      {mutation => renderSetFavButton(isFavorite, mutation, id, typeName)}
    </Mutation>
  )
}

function renderSetFavButton(isFavorite, mutation, id, typeName) {
  return (
    <SetFavoriteButton
      typeName={typeName}
      isFavorite={isFavorite}
      onSubmit={() => {
        return mutation({
          variables: {
            input: {
              target: { id, type: typeName },
              favorite: !isFavorite,
            },
          },
        })
      }}
    />
  )
}

QuerySetFavoriteButton.propTypes = {
  id: oneOfShape({ serviceID: p.string, rotationID: p.string }),
}
