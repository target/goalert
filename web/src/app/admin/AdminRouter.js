import React from 'react'
import { Switch, Route } from 'react-router-dom'
import { GenericError, PageNotFound } from '../error-pages/Errors'
import AdminConfig from './AdminConfig'
import AdminLimits from './AdminLimits'
import AdminToolbox from './AdminToolbox'
import { useSessionInfo } from '../util/RequireConfig'
import Spinner from '../loading/components/Spinner'

function AdminRouter() {
  const { isAdmin, ready } = useSessionInfo()
  if (!ready) {
    return <Spinner />
  }
  if (!isAdmin) {
    return <GenericError error='Access Denied' />
  }

  return (
    <Switch>
      <Route exact path='/admin/config' component={AdminConfig} />
      <Route exact path='/admin/limits' component={AdminLimits} />
      <Route exact path='/admin/toolbox' component={AdminToolbox} />
      <Route component={PageNotFound} />
    </Switch>
  )
}

export default AdminRouter
