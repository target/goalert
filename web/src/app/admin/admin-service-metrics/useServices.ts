import _ from 'lodash'
import React, { useLayoutEffect, useEffect, useRef, useState } from 'react'
import { gql, useClient } from 'urql'
import { Service, ServiceSearchOptions } from '../../../schema'

const servicesQuery = gql`
  query services($input: ServiceSearchOptions!) {
    services(input: $input) {
      nodes {
        id
        name
        onCallUsers {
          userID
        }
        escalationPolicy {
          id
          name
          steps {
            targets {
              name
              type
            }
          }
        }
        integrationKeys {
          type
          name
        }
        heartbeatMonitors {
          name
          timeoutMinutes
          lastHeartbeat
        }
      }
      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
`

const QUERY_LIMIT = 10

export type ServiceData = {
  services: Service[]
  loading: boolean
  error: Error | undefined
}

export function useServices(
  options: ServiceSearchOptions,
  depKey: string,
  pause?: boolean,
): ServiceData {
  const [services, setServices] = useState<Service[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | undefined>()
  const key = useRef(depKey)
  key.current = depKey

  useEffect(() => {
    return () => {
      // cancel on unmount
      key.current = ''
    }
  }, [])

  const client = useClient()
  const fetch = React.useCallback(async () => {
    setServices([])
    setLoading(true)
    setError(undefined)
    if (pause) {
      return
    }
    async function fetchServices(
      cursor: string,
    ): Promise<[Service[], boolean, string, Error | undefined]> {
      const q = await client
        .query(servicesQuery, {
          input: {
            first: QUERY_LIMIT,
            after: cursor,
          },
        })
        .toPromise()

      if (q.error) {
        return [[], false, '', q.error]
      }

      return [
        q.data.services.nodes,
        q.data.services.pageInfo.hasNextPage,
        q.data.services.pageInfo.endCursor,
        undefined,
      ]
    }

    const throttledSetServices = _.throttle(
      (services, loading) => {
        setServices(services)
        setLoading(loading)
      },
      3000,
      { leading: true },
    )

    let endCursor = ''
    let hasNextPage = true
    let error = null
    let services = []
    let allServices: Service[] = []
    while (hasNextPage) {
      ;[services, hasNextPage, endCursor, error] =
        await fetchServices(endCursor)
      if (key.current !== depKey) return // abort if the key has changed
      if (error) {
        setError(error)
        throttledSetServices.cancel()
        return
      }
      allServices = allServices.concat(services)
      throttledSetServices(allServices, true)
    }

    throttledSetServices(allServices, false)
  }, [depKey, pause])

  useLayoutEffect(() => {
    fetch()
  }, [depKey, pause])

  return {
    services,
    loading,
    error,
  }
}
