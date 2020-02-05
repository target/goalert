import React from 'react'
import { Switch, Route } from 'react-router-dom'

import AlertList from './pages/AlertsIndexPage'
import AlertDetails from './pages/AlertDetailPage'
import { PageNotFound } from '../error-pages/Errors'

export default function AlertRouter() {
  return (
    <Switch>
      <Route exact path='/alerts' component={AlertList} />
      <Route exact path='/alerts/:alertID' component={AlertDetails} />
      <Route component={PageNotFound} />
    </Switch>
  )
}
