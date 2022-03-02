import React from 'react'
import { gql } from '@apollo/client'
import { Routes, Route } from 'react-router-dom'
import { PageNotFound } from '../error-pages/Errors'
import RotationDetails from './RotationDetails'
import RotationCreateDialog from './RotationCreateDialog'
import SimpleListPage from '../lists/SimpleListPage'

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

export default function RotationRouter() {
  function renderList() {
    return (
      <SimpleListPage
        query={query}
        variables={{ input: { favoritesFirst: true } }}
        mapDataNode={(n) => ({
          title: n.name,
          subText: n.description,
          url: n.id,
          isFavorite: n.isFavorite,
        })}
        createForm={<RotationCreateDialog />}
        createLabel='Rotation'
      />
    )
  }

  return (
    <Routes>
      <Route path='/' element={renderList()} />
      <Route path=':rotationID' element={<RotationDetails />} />
      <Route element={<PageNotFound />} />
    </Routes>
  )
}
