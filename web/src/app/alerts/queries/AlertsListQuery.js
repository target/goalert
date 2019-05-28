import gql from 'graphql-tag'

export const alertsQuery = gql`
  query alerts(
    $favorite_services_only: Boolean!
    $service_id: String
    $search: String
    $sort_desc: Boolean!
    $limit: Int!
    $offset: Int!
    $omit_active: Boolean!
    $omit_closed: Boolean!
    $omit_triggered: Boolean!
    $sort_by: AlertSortBy
    $favorites_first: Boolean!
    $favorites_only: Boolean!
    $services_limit: Int!
    $services_search: String!
  ) {
    alerts2(
      options: {
        favorite_services_only: $favorite_services_only
        service_id: $service_id
        search: $search
        sort_desc: $sort_desc
        limit: $limit
        offset: $offset
        omit_active: $omit_active
        omit_closed: $omit_closed
        omit_triggered: $omit_triggered
        sort_by: $sort_by
      }
    ) {
      total_count
      items {
        number: _id
        id
        status: status_2
        created_at
        summary
        service {
          id
          name
        }
      }
    }

    services2(
      options: {
        favorites_first: $favorites_first
        favorites_only: $favorites_only
        limit: $services_limit
        search: $services_search
      }
    ) {
      items {
        id
      }
    }
  }
`
