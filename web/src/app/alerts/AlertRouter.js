import React from 'react'
import { Routes, Route } from 'react-router-dom'

import AlertDetails from './pages/AlertDetailPage'
import { PageNotFound } from '../error-pages/Errors'
import AlertsList from './AlertsList'

export default function AlertRouter() {
  return (
    <Routes>
      <Route exact path='/alerts' component={AlertsList} />
      <Route exact path='/alerts/:alertID' component={AlertDetails} />
      <Route component={PageNotFound} />
    </Routes>
  )
}
