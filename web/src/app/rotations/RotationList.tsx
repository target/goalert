import React from 'react'
import { gql } from 'urql'
import QueryList from '../lists/QueryList'
import RotationCreateDialog from './RotationCreateDialog'

const query = gql`
  query rotationsQuery($input: RotationSearchOptions) {
    data: rotations(input: $input) {
      nodes {
        id
        name
        description
        isFavorite
      }
      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
`

export default function RotationList(): JSX.Element {
  return (
    <QueryList
      query={query}
      variables={{ input: { favoritesFirst: true } }}
      mapDataNode={(n) => ({
        title: n.name,
        subText: n.description,
        url: n.id,
        isFavorite: n.isFavorite,
      })}
      renderCreateDialog={(onClose) => (
        <RotationCreateDialog onClose={onClose} />
      )}
      createLabel='Rotation'
    />
  )
}
