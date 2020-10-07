import gql from 'graphql-tag'
import _ from 'lodash-es'
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
  query($id: ID!) {
    user(id: $id) {
      id
      name
    }
  }
`

// useUserInfo will add `user` info to array items that contain a `userID`.
export function useUserInfo<T extends HasUserID>(
  items: T[],
): (T & WithUserInfo)[] {
  const variables = _.uniq(
    items.map((item) => item.userID).sort(),
  ).map((id) => ({ id }))

  const { data, loading, error } = useMultiQuery(infoQuery, {
    variables,
    fetchPolicy: 'cache-first',
    pollInterval: 0,
    skip: items.length === 0,
  })

  if (error && !loading) {
    return items.map((item) => ({
      ...item,
      user: { id: item.userID, name: 'Error: ' + error.message },
    }))
  }
  if (!data) {
    return items.map((item) => ({
      ...item,
      user: { id: item.userID, name: 'Loading...' },
    }))
  }

  const lookup: Record<string, string> = {}
  data.forEach((res: WithUserInfo) => {
    lookup[res.user.id] = res.user.name
  })

  return items.map((item: T) => ({
    ...item,
    user: { id: item.userID, name: lookup[item.userID] || 'Unknown User' },
  }))
}
