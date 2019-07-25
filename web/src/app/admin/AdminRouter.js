import React from 'react'
import { Switch, Route } from 'react-router-dom'
import { GenericError, PageNotFound } from '../error-pages/Errors'
import AdminConfig from './AdminConfig'
import RequireConfig from '../util/RequireConfig'

export const AdminRouter = () => (
  <Switch>
    <Route
      exact
      path='/admin/config'
      render={() => (
        <RequireConfig isAdmin else={<GenericError error='Access Denied' />}>
          <AdminConfig />
        </RequireConfig>
      )}
    />
    <Route component={PageNotFound} />
  </Switch>
)

export default AdminRouter
