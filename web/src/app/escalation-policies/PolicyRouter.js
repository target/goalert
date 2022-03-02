import React from 'react'
import { gql } from '@apollo/client'
import { Routes, Route } from 'react-router-dom'
import PolicyCreateDialog from './PolicyCreateDialog'
import PolicyDetails from './PolicyDetails'
import PolicyServicesQuery from './PolicyServicesQuery'
import { PageNotFound } from '../error-pages/Errors'
import SimpleListPage from '../lists/SimpleListPage'

const query = gql`
  query epsQuery($input: EscalationPolicySearchOptions) {
    data: escalationPolicies(input: $input) {
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

export default function PolicyRouter() {
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
        createForm={<PolicyCreateDialog />}
        createLabel='Escalation Policy'
      />
    )
  }

  return (
    <Routes>
      <Route path='/' element={renderList()} />
      <Route path=':escalationPolicyID' element={<PolicyDetails />} />
      <Route
        path=':escalationPolicyID/services'
        element={<PolicyServicesQuery />}
      />
      <Route element={<PageNotFound />} />
    </Routes>
  )
}
