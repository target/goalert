import React from 'react'
import { Routes, Route } from 'react-router-dom'

import AlertDetails from './pages/AlertDetailPage'
import { PageNotFound } from '../error-pages/Errors'
import AlertsList from './AlertsList'

export default function AlertRouter() {
  return (
    <Routes>
      <Route path='/' element={<AlertsList />} />
      <Route path=':alertID' element={<AlertDetails />} />
      <Route element={<PageNotFound />} />
    </Routes>
  )
}
