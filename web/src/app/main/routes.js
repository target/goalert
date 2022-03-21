import React from 'react'
import { Navigate, Route } from 'react-router-dom'
import joinURL from '../util/joinURL'
import RotationRouter from '../rotations/RotationRouter'
import AlertRouter from '../alerts/AlertRouter'
import ScheduleRouter from '../schedules/ScheduleRouter'
import PolicyRouter from '../escalation-policies/PolicyRouter'
import ServiceRouter from '../services/ServiceRouter'
import UserRouter, { ProfileRouter } from '../users/UserRouter'
import AdminRouter from '../admin/AdminRouter'
import WizardRouter from '../wizard/WizardRouter'
import Documentation from '../documentation/Documentation'

export const getPath = (p) => (Array.isArray(p.path) ? p.path[0] : p.path)

export function renderRoutes(routeConfig = []) {
  const routes = []

  routeConfig.forEach((cfg, idx) => {
    const _path = cfg.path
    const path = Array.isArray(_path) ? _path[0] : _path

    if (Array.isArray(_path)) {
      // add alias routes (for compatibility)
      _path.slice(1).forEach((p, pIdx) => {
        routes.push(
          <Route
            key={`alias_${idx}_${pIdx}`}
            path={p}
            element={<Navigate to={path.replace('/*', '')} replace />}
          />,
        )
        if (p !== '/*') {
          // redirect nested paths (e.g. /on_call_schedules/foo to /schedules/foo)
          routes.push(
            <Route
              key={`alias_${idx}_${pIdx}_splat`}
              path={joinURL(p, '*')}
              element={<Navigate replace to={joinURL(path, '*')} />}
            />,
          )
        }
      })
    }

    if (cfg.subRoutes && cfg.subRoutes.length) {
      routes.push(
        <Route
          key={`alias_${idx}`}
          path={path.replace('/*', '')}
          element={
            <Navigate
              to={path.replace('/*', '') + cfg.subRoutes[0].path}
              replace
            />
          }
        />,
      )
    }

    routes.push(
      <Route key={'route_' + idx} element={<cfg.element />} path={path} />,
    )
  })

  return routes
}

// used by new app and the toolbar title
export default [
  {
    title: 'Alerts',
    path: ['/alerts/*', '/'],
    element: AlertRouter,
  },
  {
    title: 'Rotations',
    path: '/rotations/*',
    element: RotationRouter,
  },
  {
    title: 'Schedules',
    path: ['/schedules/*', '/on_call_schedules/*'],
    element: ScheduleRouter,
  },
  {
    title: 'Escalation Policies',
    path: ['/escalation-policies/*', '/escalation_policies/*'],
    element: PolicyRouter,
  },
  {
    title: 'Services',
    path: '/services/*',
    element: ServiceRouter,
  },
  {
    title: 'Users',
    path: '/users/*',
    element: UserRouter,
  },
  {
    nav: false,
    title: 'Setup Wizard',
    path: '/wizard/*',
    element: WizardRouter,
  },
  {
    nav: false,
    title: 'Profile',
    path: '/profile/*',
    element: ProfileRouter,
  },
  {
    nav: false,
    title: 'Admin',
    path: '/admin/*',
    element: AdminRouter,
    subRoutes: [
      {
        title: 'Config',
        path: '/config',
        element: AdminRouter,
      },
      {
        title: 'Message Logs',
        path: '/message-logs',
        element: AdminRouter,
      },
      {
        title: 'Switchover',
        path: '/switchover',
        element: AdminRouter,
      },
      {
        title: 'System Limits',
        path: '/limits',
        element: AdminRouter,
      },
      {
        title: 'Toolbox',
        path: '/toolbox',
        element: AdminRouter,
      },
    ],
  },
  {
    nav: false,
    title: 'Documentation',
    path: '/docs',
    element: Documentation,
  },
]
