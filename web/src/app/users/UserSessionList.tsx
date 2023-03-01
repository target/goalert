import React, { useState } from 'react'
import FlatList from '../lists/FlatList'
import {
  QueryHookOptions,
  useMutation,
  useQuery,
  ApolloError,
  gql,
} from '@apollo/client'
import { Button, Card, Grid, IconButton } from '@mui/material'
import DeleteIcon from '@mui/icons-material/Delete'
import { UserSession } from '../../schema'
import Bowser from 'bowser'
import _ from 'lodash'
import FormDialog from '../dialogs/FormDialog'
import { nonFieldErrors } from '../util/errutil'
import { Time } from '../util/Time'

const profileQuery = gql`
  query {
    user {
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

const byUserQuery = gql`
  query ($userID: ID!) {
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

const mutationLogoutOne = gql`
  mutation ($id: ID!) {
    deleteAll(input: [{ id: $id, type: userSession }])
  }
`

const mutationLogoutAll = gql`
  mutation {
    endAllAuthSessionsByCurrentUser
  }
`

function friendlyUAString(ua: string): string {
  if (!ua) return 'Unknown device'
  const b = Bowser.getParser(ua)

  let str
  if (b.getBrowserName()) {
    str = b.getBrowserName() + ' ' + b.getBrowserVersion().split('.')[0]
  }
  if (!str) {
    str = 'Unknown device'
  }

  if (b.getOSName()) {
    str += ' on ' + b.getOSName()
  }

  if (b.getPlatformType()) {
    str += ' (' + b.getPlatformType() + ')'
  }

  return str
}

type Session = {
  id: string
  userAgent: string
}

export type UserSessionListProps = {
  userID: string
}

export default function UserSessionList({
  userID,
}: UserSessionListProps): JSX.Element {
  // handles both logout all and logout individual sessions
  const [endSession, setEndSession] = useState<Session | 'all' | null>(null)

  const options: QueryHookOptions = {}
  if (userID) {
    options.variables = { userID }
  }
  const { data } = useQuery(userID ? byUserQuery : profileQuery, options)

  const sessions: UserSession[] = _.sortBy(
    data?.user?.sessions || [],
    (s: UserSession) => (s.current ? '_' + s.lastAccessAt : s.lastAccessAt),
  ).reverse()

  const [logoutOne, logoutOneStatus] = useMutation(mutationLogoutOne, {
    variables: { id: (endSession as Session)?.id },
    onCompleted: () => setEndSession(null),
  })
  const [logoutAll, logoutAllStatus] = useMutation(mutationLogoutAll, {
    onCompleted: () => setEndSession(null),
  })

  return (
    <React.Fragment>
      <Grid container spacing={2}>
        {!userID && (
          <Grid item xs={12} container justifyContent='flex-end'>
            <Button
              variant='outlined'
              data-cy='reset'
              onClick={() => setEndSession('all')}
            >
              Log Out Other Sessions
            </Button>
          </Grid>
        )}
        <Grid item xs={12}>
          <Card>
            <FlatList
              emptyMessage='No active sessions'
              items={sessions.map((s) => ({
                title: friendlyUAString(s.userAgent),
                highlight: s.current,
                secondaryAction: s.current ? null : (
                  <IconButton
                    onClick={() =>
                      setEndSession({
                        id: s.id,
                        userAgent: s.userAgent,
                      })
                    }
                    size='large'
                  >
                    <DeleteIcon />
                  </IconButton>
                ),
                subText: (
                  <Time
                    format='relative'
                    time={s.lastAccessAt}
                    prefix='Last access: '
                    min={{ minutes: 2 }}
                  />
                ),
              }))}
            />
          </Card>
        </Grid>
      </Grid>

      {endSession === 'all' && (
        <FormDialog
          title='Are you sure?'
          confirm
          loading={logoutAllStatus.loading}
          errors={nonFieldErrors(logoutAllStatus.error as ApolloError)}
          subTitle='This will log you out of all other sessions.'
          onSubmit={() => logoutAll()}
          onClose={() => setEndSession(null)}
        />
      )}

      {endSession && endSession !== 'all' && (
        <FormDialog
          title='Are you sure?'
          confirm
          loading={logoutOneStatus.loading}
          errors={nonFieldErrors(logoutOneStatus.error as ApolloError)}
          subTitle={`This will log you out of your "${friendlyUAString(
            endSession.userAgent,
          )}" session.`}
          onSubmit={() => logoutOne()}
          onClose={() => setEndSession(null)}
        />
      )}
    </React.Fragment>
  )
}
