import React from 'react'
import p from 'prop-types'
import gql from 'graphql-tag'
import { useQuery, useMutation } from 'react-apollo'
import { SetFavoriteButton } from './SetFavoriteButton'
import { oneOfShape } from '../util/propTypes'

const queries = {
  service: gql`
    query serviceFavQuery($id: ID!) {
      data: service(id: $id) {
        id
        isFavorite
      }
    }
  `,
  rotation: gql`
    query rotationFavQuery($id: ID!) {
      data: rotation(id: $id) {
        id
        isFavorite
      }
    }
  `,
  schedule: gql`
    query scheduleFavQuery($id: ID!) {
      data: schedule(id: $id) {
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
  let id, typeName
  if (props.rotationID) {
    typeName = 'rotation'
    id = props.rotationID
  } else if (props.serviceID) {
    typeName = 'service'
    id = props.serviceID
  } else if (props.scheduleID) {
    typeName = 'schedule'
    id = props.scheduleID
  } else {
    throw new Error('unknown type')
  }
  const { data, loading } = useQuery(queries[typeName], {
    variables: { id },
  })
  const isFavorite = data && data.data && data.data.isFavorite
  const [toggleFav] = useMutation(mutation, {
    variables: {
      input: { target: { id, type: typeName }, favorite: !isFavorite },
    },
  })

  return (
    <SetFavoriteButton
      typeName={typeName}
      isFavorite={isFavorite}
      loading={!data && loading}
      onClick={() => toggleFav()}
      type={props.type}
    />
  )
}

QuerySetFavoriteButton.propTypes = {
  id: oneOfShape({
    serviceID: p.string,
    rotationID: p.string,
    scheduleID: p.string,
  }),

  // todo: extend SetFavoriteButtonProps when converting to ts
  type: p.oneOf(['service']),
}
