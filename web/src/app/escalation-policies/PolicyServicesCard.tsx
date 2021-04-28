import React from 'react'
import Card from '@material-ui/core/Card'
import { makeStyles } from '@material-ui/core/styles'
import FlatList, { FlatListListItem } from '../lists/FlatList'
import _ from 'lodash'

const useStyles = makeStyles(() => ({
  card: {
    width: '100%',
  },
}))

interface PolicyServicesCardProps {
  services: { id: string; name: string }[]
}

function PolicyServicesCard(props: PolicyServicesCardProps): JSX.Element {
  const classes = useStyles()

  function getServicesItems(): FlatListListItem[] {
    const items = props.services.map((service) => ({
      title: service.name,
      url: `/services/${service.id}`,
    }))

    // case-insensitive sort
    return _.sortBy(items, (i) => i.title.toLowerCase(), ['title'])
  }

  return (
    <Card className={classes.card}>
      <FlatList
        emptyMessage='No services are associated with this Escalation Policy.'
        items={getServicesItems()}
      />
    </Card>
  )
}

export default PolicyServicesCard
