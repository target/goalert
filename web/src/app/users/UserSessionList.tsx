import React, { useState } from 'react'
import FlatList from '../lists/FlatList'
import { useMutation, useQuery, CombinedError, gql } from 'urql'
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
}: UserSessionListProps): React.JSX.Element {
  // handles both logout all and logout individual sessions
  const [endSession, setEndSession] = useState<Session | 'all' | null>(null)

  let variables = {}
  if (userID) {
    variables = { userID }
  }
  const [{ data }] = useQuery({
    query: userID ? byUserQuery : profileQuery,
    variables,
  })

  const sessions: UserSession[] = _.sortBy(
    data?.user?.sessions || [],
    (s: UserSession) => (s.current ? '_' + s.lastAccessAt : s.lastAccessAt),
  ).reverse()

  const [logoutOneStatus, logoutOne] = useMutation(mutationLogoutOne)
  const [logoutAllStatus, logoutAll] = useMutation(mutationLogoutAll)

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
          loading={logoutAllStatus.fetching}
          errors={nonFieldErrors(logoutAllStatus.error as CombinedError)}
          subTitle='This will log you out of all other sessions.'
          onSubmit={() =>
            logoutAll().then((result) => {
              if (!result.error) setEndSession(null)
            })
          }
          onClose={() => setEndSession(null)}
        />
      )}

      {endSession && endSession !== 'all' && (
        <FormDialog
          title='Are you sure?'
          confirm
          loading={logoutOneStatus.fetching}
          errors={nonFieldErrors(logoutOneStatus.error as CombinedError)}
          subTitle={`This will log you out of your "${friendlyUAString(
            endSession.userAgent,
          )}" session.`}
          onSubmit={() =>
            logoutOne({ id: (endSession as Session)?.id }).then((result) => {
              if (!result.error) setEndSession(null)
            })
          }
          onClose={() => setEndSession(null)}
        />
      )}
    </React.Fragment>
  )
}
