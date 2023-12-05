import React from 'react'
import { gql, useQuery } from 'urql'
import { PropTypes as p } from 'prop-types'
import Card from '@mui/material/Card'
import CardHeader from '@mui/material/CardHeader'
import { UserAvatar } from '../util/avatars'
import makeStyles from '@mui/styles/makeStyles'
import { styles as globalStyles } from '../styles/materialStyles'
import FlatList from '../lists/FlatList'
import { Error } from '@mui/icons-material'
import _ from 'lodash'
import { Warning } from '../icons'

const useStyles = makeStyles((theme) => {
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

const stepText = (s) => {
  return `Step #${s + 1}`
}
const stepLengthText = (s) => {
  if (s > 0) {
    if (s === 1) {
      return `${s + 1} user assigned`
    }
    return `${s + 1} users assigned`
  }
  return 'No users assigned'
}
export default function ServiceOnCallList({ serviceID }) {
  const classes = useStyles()
  const [{ data, error }] = useQuery({
    query,
    variables: { id: serviceID },
  })

  let items = []
  let sections = []
  const style = {}
  if (error) {
    items = [
      {
        title: 'Error: ' + error.message,
        icon: <Error />,
      },
    ]
    style.color = 'gray'
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
ServiceOnCallList.propTypes = {
  serviceID: p.string.isRequired,
}
