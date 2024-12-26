import React, { Suspense, useState } from 'react'
import { gql, useQuery } from 'urql'
import { useURLParam } from '../actions'
import getServiceFilters from '../util/getServiceFilters'
import ServiceCreateDialog from './ServiceCreateDialog'
import ServiceFilterContainer from './ServiceFilterContainer'
import { ServiceConnection } from '../../schema'
import ListPageControls from '../lists/ListPageControls'
import Search from '../util/Search'
import FlatList from '../lists/FlatList'
import { FavoriteIcon } from '../util/SetFavoriteButton'

const query = gql`
  query servicesQuery($input: ServiceSearchOptions) {
    services(input: $input) {
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
const context = { suspense: false }
export default function ServiceList(): React.JSX.Element {
  const [search, setSearch] = useURLParam<string>('search', '')
  const { labelKey, labelValue, integrationKey } = getServiceFilters(search)
  const [create, setCreate] = useState(false)
  const [cursor, setCursor] = useState('')

  const inputVars = {
    favoritesFirst: true,
    search,
    after: cursor,
  }

  const [q] = useQuery<{ services: ServiceConnection }>({
    query,
    variables: { input: inputVars },
    context,
  })
  const nextCursor = q.data?.services.pageInfo.hasNextPage
    ? q.data?.services.pageInfo.endCursor
    : ''
  // cache the next page
  useQuery({
    query,
    variables: { input: { ...inputVars, after: nextCursor } },
    context,
    pause: !nextCursor,
  })

  return (
    <React.Fragment>
      <Suspense>
        {create && <ServiceCreateDialog onClose={() => setCreate(false)} />}
      </Suspense>
      <ListPageControls
        createLabel='Service'
        nextCursor={nextCursor}
        onCursorChange={setCursor}
        loading={q.fetching}
        onCreateClick={() => setCreate(true)}
        slots={{
          search: (
            <Search
              endAdornment={
                <ServiceFilterContainer
                  value={{ labelKey, labelValue, integrationKey }}
                  onChange={({ labelKey, labelValue, integrationKey }) => {
                    const labelSearch = labelKey
                      ? labelKey + '=' + labelValue
                      : ''
                    const intKeySearch = integrationKey
                      ? 'token=' + integrationKey
                      : ''
                    const searchStr =
                      intKeySearch && labelSearch
                        ? intKeySearch + ' ' + labelSearch
                        : intKeySearch + labelSearch
                    setSearch(searchStr)
                  }}
                  onReset={() => setSearch('')}
                />
              }
            />
          ),
          list: (
            <FlatList
              emptyMessage='No results'
              items={
                q.data?.services.nodes.map((u) => ({
                  title: u.name,
                  subText: u.description,
                  url: u.id,
                  secondaryAction: u.isFavorite ? <FavoriteIcon /> : undefined,
                })) || []
              }
            />
          ),
        }}
      />
    </React.Fragment>
  )
}
