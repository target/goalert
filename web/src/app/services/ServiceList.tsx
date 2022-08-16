import React from 'react'
import { gql } from 'urql'
import { useURLParam } from '../actions'
import SimpleListPage from '../lists/SimpleListPage'
import getServiceLabel from '../util/getServiceLabel'
import ServiceCreateDialog from './ServiceCreateDialog'
import ServiceFilterContainer from './ServiceFilterContainer'

const query = gql`
  query servicesQuery($input: ServiceSearchOptions) {
    data: services(input: $input) {
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

export default function ServiceList(): JSX.Element {
  const [searchParam, setSearchParam] = useURLParam<string>('search', '')
  const { labelKey, labelValue } = getServiceLabel(searchParam)

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
      createForm={<ServiceCreateDialog />}
      createLabel='Service'
      searchAdornment={
        <ServiceFilterContainer
          value={{ labelKey, labelValue }}
          onChange={({ labelKey, labelValue }) =>
            setSearchParam(labelKey ? labelKey + '=' + labelValue : '')
          }
          onReset={() => setSearchParam('')}
        />
      }
    />
  )
}
