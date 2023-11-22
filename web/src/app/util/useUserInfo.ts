import { gql } from 'urql'
import _ from 'lodash'
import useMultiQuery from './useMultiQuery'

interface HasUserID {
  userID: string
}

export interface WithUserInfo {
  user: {
    id: string
    name: string
  }
}

const infoQuery = gql`
  query ($id: ID!) {
    user(id: $id) {
      id
      name
    }
  }
`

const noSuspense = { suspense: false }

// useUserInfo will add `user` info to array items that contain a `userID`.
export function useUserInfo<T extends HasUserID>(
  items: T[],
): (T & WithUserInfo)[] {
  const variables = _.uniq(items.map((item) => item.userID).sort()).map(
    (id) => ({ id }),
  )

  const [{ data, fetching, error }] = useMultiQuery({
    query: infoQuery,
    variables,
    requestPolicy: 'cache-first',
    pause: items.length === 0,
    context: noSuspense,
  })

  // handle error
  if (error && !fetching) {
    return items.map((item) => ({
      ...item,
      user: { id: item.userID, name: 'Error: ' + error.message },
    }))
  }

  // handle none loaded
  if (!data) {
    return items.map((item) => ({
      ...item,
      user: { id: item.userID, name: 'Loading...' },
    }))
  }

  // handle some loaded
  const lookup: Record<string, string> = {}
  data.forEach((res: WithUserInfo) => {
    if (res?.user) {
      lookup[res.user.id] = res.user.name
    }
  })

  return items.map((item: T) => ({
    ...item,
    user: { id: item.userID, name: lookup[item.userID] || 'Unknown User' },
  }))
}
