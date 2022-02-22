import React from 'react'
import { Routes, Route } from 'react-router-dom'
import { GenericError, PageNotFound } from '../error-pages/Errors'
import AdminConfig from './AdminConfig'
import AdminLimits from './AdminLimits'
import AdminToolbox from './AdminToolbox'
import AdminDebugMessagesLayout from './admin-message-logs/AdminDebugMessagesLayout'
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
    <Routes>
      <Route exact path='/admin/config' component={AdminConfig} />
      <Route exact path='/admin/limits' component={AdminLimits} />
      <Route exact path='/admin/toolbox' component={AdminToolbox} />
      <Route
        exact
        path='/admin/message-logs'
        component={AdminDebugMessagesLayout}
      />

      <Route component={PageNotFound} />
    </Routes>
  )
}

export default AdminRouter
