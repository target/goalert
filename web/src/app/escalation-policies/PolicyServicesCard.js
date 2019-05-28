import React, { Component } from 'react'
import { PropTypes as p } from 'prop-types'
import Card from '@material-ui/core/Card'
import withStyles from '@material-ui/core/styles/withStyles'
import FlatList from '../lists/FlatList'

const styles = {
  card: {
    width: '100%',
  },
}

@withStyles(styles)
export default class PolicyServicesCard extends Component {
  static propTypes = {
    services: p.arrayOf(
      p.shape({
        id: p.string.isRequired,
        name: p.string.isRequired,
      }),
    ).isRequired,
  }

  getServicesItems = () => {
    return this.props.services.map(service => ({
      title: service.name,
      url: `/services/${service.id}`,
    }))
  }

  render() {
    return (
      <Card className={this.props.classes.card}>
        <FlatList
          emptyMessage='No services are associated with this Escalation Policy.'
          items={this.getServicesItems()}
        />
      </Card>
    )
  }
}
