import React from 'react'
import { gql, useQuery } from '@apollo/client'
import { PropTypes as p } from 'prop-types'
import Card from '@mui/material/Card'
import CardHeader from '@mui/material/CardHeader'
import { UserAvatar } from '../util/avatars'
import { CircularProgress } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import { styles as globalStyles } from '../styles/materialStyles'
import FlatList from '../lists/FlatList'
import { Error } from '@mui/icons-material'
import _ from 'lodash'

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

export default function ServiceOnCallList({ serviceID }) {
  const classes = useStyles()
  const { data, loading, error } = useQuery(query, {
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
  } else if (!data && loading) {
    items = [
      {
        title: 'Fetching users...',
        icon: <CircularProgress />,
      },
    ]
    style.color = 'gray'
    sections = [
      {
        title: 'Fetching users...',
        icon: <CircularProgress />,
      },
    ]
  } else {
    const chainedItems = _.chain(data?.service?.onCallUsers)
    sections = chainedItems
      .groupBy('stepNumber')
      .keys()
      .map((s) => ({ title: stepText(Number(s)) }))
      .value()

    items = _.chain(data?.service?.onCallUsers)
      .sortBy(['stepNumber', 'userName'])
      .map((u) => ({
        title: u.userName,
        subText: stepText(u.stepNumber),
        icon: <UserAvatar userID={u.userID} />,
        section: stepText(u.stepNumber),
        url: `/users/${u.userID}`,
      }))
      .value()
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
