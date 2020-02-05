import React from 'react'
import { Switch, Route } from 'react-router-dom'

import AlertDetails from './pages/AlertDetailPage'
import { PageNotFound } from '../error-pages/Errors'
import PageActions from '../util/PageActions'
import AlertsListFilter from './components/AlertsListFilter'
import Search from '../util/Search'
import AlertsList from './components/AlertsList'

export default function AlertRouter() {
  return (
    <Switch>
      <Route
        exact
        path='/alerts'
        render={() => (
          <React.Fragment>
            <PageActions>
              <AlertsListFilter key='filter' />
              <Search key='search' />
            </PageActions>
            <AlertsList />
          </React.Fragment>
        )}
      />
      <Route exact path='/alerts/:alertID' component={AlertDetails} />
      <Route component={PageNotFound} />
    </Switch>
  )
}
