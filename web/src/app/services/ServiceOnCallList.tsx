import React from 'react'
import { gql, useQuery } from 'urql'
import Card from '@mui/material/Card'
import CardHeader from '@mui/material/CardHeader'
import { UserAvatar } from '../util/avatars'
import makeStyles from '@mui/styles/makeStyles'
import { styles as globalStyles } from '../styles/materialStyles'
import { Error } from '@mui/icons-material'
import _ from 'lodash'
import { Warning } from '../icons'
import { Theme } from '@mui/material'
import CompList from '../lists/CompList'
import {
  CompListItemNav,
  CompListItemText,
  CompListSection,
} from '../lists/CompListItems'
import { ServiceOnCallUser } from '../../schema'

const useStyles = makeStyles((theme: Theme) => {
  const { cardHeader } = globalStyles(theme)

  return {
    cardHeader,
  }
})

const query = gql`
  query onCallQuery($id: ID!) {
    service(id: $id) {
      id
      escalationPolicy {
        id
        name
        steps {
          stepNumber
        }
      }
      onCallUsers {
        userID
        userName
        stepNumber
      }
    }
  }
`

const stepText = (s: number): string => {
  return `Step #${s + 1}`
}
const stepLengthText = (s: number): string => {
  if (s > 0) {
    if (s === 1) {
      return `${s} user assigned`
    }
    return `${s} users assigned`
  }
  return 'No users assigned'
}
export default function ServiceOnCallList({
  serviceID,
}: {
  serviceID: string
}): React.ReactNode {
  const classes = useStyles()
  const [{ data, error }] = useQuery({
    query,
    variables: { id: serviceID },
  })

  let sections: {
    title: string
    subText: string
    icon: React.ReactNode
    users: ServiceOnCallUser[]
  }[] = []

  const chainedSteps = _.chain(data?.service?.escalationPolicy?.steps)
  const sortedItems = _.chain(data?.service?.onCallUsers)
    .sortBy(['stepNumber', 'userName'])
    .value()

  sections = chainedSteps
    .groupBy('stepNumber')
    .keys()
    .map((s) => {
      const usersAssigned = sortedItems.filter(
        (item) => item.stepNumber === Number(s),
      ).length
      return {
        title: stepText(Number(s)),
        subText: stepLengthText(usersAssigned),
        icon: usersAssigned === 0 && (
          <Warning message='No user assigned for step.' />
        ),
        users: sortedItems.filter((item) => item.stepNumber === Number(s)),
      }
    })
    .value()

  return (
    <Card>
      <CardHeader
        className={classes.cardHeader}
        component='h3'
        title='On Call Users'
      />
      <CompList emptyMessage='No users on-call for this service'>
        {error && (
          <CompListItemText
            title='Error'
            icon={<Error />}
            subText={error.message}
          />
        )}
        {sections.map((s, i) => {
          return (
            <CompListSection
              defaultOpen={i === 0}
              title={s.title}
              subText={s.subText}
              icon={s.icon}
            >
              {s.users.map((u) => (
                <CompListItemNav
                  title={u.userName}
                  icon={<UserAvatar userID={u.userID} />}
                  url={`/users/${u.userID}`}
                />
              ))}
            </CompListSection>
          )
        })}
      </CompList>
    </Card>
  )
}
