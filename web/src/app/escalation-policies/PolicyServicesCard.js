import React from 'react'
import { PropTypes as p } from 'prop-types'
import Card from '@material-ui/core/Card'
import { makeStyles } from '@material-ui/core/styles'
import FlatList from '../lists/FlatList'

const useStyles = makeStyles(() => ({
  card: {
    width: '100%',
  },
}))

function PolicyServicesCard(props) {
  const classes = useStyles()

  function getServicesItems() {
    return props.services.map((service) => ({
      title: service.name,
      url: `/services/${service.id}`,
    }))
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

PolicyServicesCard.propTypes = {
  services: p.arrayOf(
    p.shape({
      id: p.string.isRequired,
      name: p.string.isRequired,
    }),
  ).isRequired,
}

export default PolicyServicesCard
