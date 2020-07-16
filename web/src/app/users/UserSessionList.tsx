import React from 'react'
import FlatList from '../lists/FlatList'
import { useSessionInfo } from '../util/RequireConfig'
import gql from 'graphql-tag'
import { useQuery } from 'react-apollo'
import { Card, ListItemText } from '@material-ui/core'
import { UserSession } from '../../schema'
import Bowser from 'bowser'
import { formatTimeSince } from '../util/timeFormat'
import _ from 'lodash-es'

const query = gql`
  query($userID: ID!) {
    user(id: $userID) {
      id
      sessions {
        id
        userAgent
        current
        createdAt
        lastAccessAt
      }
    }
  }
`

export interface UserSessionListProps {
  userID?: string
}

function friendlyUAString(ua: string): string {
  const b = Bowser.getParser(ua)

  return `${b.getBrowserName()} ${
    b.getBrowserVersion().split('.')[0]
  } on ${b.getOSName()} (${b.getPlatformType()})`
}

export default function UserSessionList(
  props: UserSessionListProps,
): ReactElement {
  const { userID: curUserID } = useSessionInfo() as any
  const userID = props.userID || curUserID
  const { data, loading } = useQuery(query, { variables: { userID } })

  const sessions: UserSession[] = _.sortBy(
    data?.user?.sessions || [],
    (s: UserSession) => (s.current ? '_' + s.lastAccessAt : s.lastAccessAt),
  )

  return (
    <Card>
      <FlatList
        loading={loading}
        items={sessions.map((s) => ({
          title: friendlyUAString(s.userAgent),
          highlight: s.current,
          secondaryAction: (
            <ListItemText>{formatTimeSince(s.lastAccessAt)}</ListItemText>
          ),
          subText: `Logged in ${formatTimeSince(s.createdAt)} ago.`,
        }))}
      />
    </Card>
  )
}
