import React from 'react'
import { gql, useQuery } from 'urql'
import Card from '@mui/material/Card'
import CardHeader from '@mui/material/CardHeader'
import { UserAvatar } from '../util/avatars'
import makeStyles from '@mui/styles/makeStyles'
import { styles as globalStyles } from '../styles/materialStyles'
import FlatList, { SectionTitle } from '../lists/FlatList'
import { Error } from '@mui/icons-material'
import _ from 'lodash'
import { Warning } from '../icons'
import { Theme } from '@mui/material'

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
          targets {
            type
          }
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

  let items = []
  let sections: SectionTitle[] = []
  if (error) {
    items = [
      {
        title: 'Error: ' + error.message,
        icon: <Error />,
      },
    ]
  } else {
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
          icon: usersAssigned === 0 && <Warning />,
        }
      })
      .value()

    items = sortedItems.map((u) => ({
      title: u.userName,
      icon: <UserAvatar userID={u.userID} />,
      section: stepText(u.stepNumber),
      url: `/users/${u.userID}`,
    }))
  }

  return (
    <Card>
      <CardHeader
        className={classes.cardHeader}
        component='h3'
        title='On Call Users'
      />
      <FlatList
        emptyMessage='No users on-call for this service'
        items={items}
        sections={sections}
        collapsable
      />
    </Card>
  )
}
