import React, { Component } from 'react'
import AlertsList from '../components/AlertsList'
import PageActions from '../../util/PageActions'

import AlertsListFilter from '../components/AlertsListFilter'
import Search from '../../util/Search'

export default class AlertsIndexPage extends Component {
  render() {
    return (
      <React.Fragment>
        <PageActions>
          <Search key='search' />
          <AlertsListFilter key='filter' />
        </PageActions>
        <AlertsList />
      </React.Fragment>
    )
  }
}
