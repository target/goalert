import React from 'react'
import { PropTypes as p } from 'prop-types'
import Card from '@material-ui/core/Card'
import CardHeader from '@material-ui/core/CardHeader'
import { UserAvatar } from '../util/avatar'
import { CircularProgress, makeStyles } from '@material-ui/core'
import { styles as globalStyles } from '../styles/materialStyles'
import FlatList from '../lists/FlatList'
import { useQuery } from 'react-apollo'
import { Error } from '@material-ui/icons'
import gql from 'graphql-tag'
import _ from 'lodash-es'

const useStyles = makeStyles(theme => {
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

const stepsText = _steps => {
  const steps = _.chain(_steps)
    .sort()
    .map(s => `#${s + 1}`)
    .value()
  if (steps.length === 1) {
    return 'Step ' + steps[0]
  }

  const last = steps.pop()
  return (
    `Steps ` + steps.join(', ') + (steps.length > 2 ? ', and ' : ' and ') + last
  )
}

export default function ServiceOnCallList({ serviceID }) {
  const classes = useStyles()
  const { data, loading, error } = useQuery(query, {
    variables: { id: serviceID },
  })

  let items = []
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
  } else {
    items = _.chain(data.service.onCallUsers)
      .groupBy('userID')
      .mapValues(v => ({
        id: v[0].userID,
        name: v[0].userName,
        steps: _.map(v, 'stepNumber'),
      }))
      .values()
      .sortBy('name')
      .map(u => ({
        title: u.name,
        subText: stepsText(u.steps),
        icon: <UserAvatar userID={u.id} />,
        url: `/users/${u.id}`,
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
      />
    </Card>
  )
}
ServiceOnCallList.propTypes = {
  serviceID: p.string.isRequired,
}
