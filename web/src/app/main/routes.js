import React from 'react'
import { Redirect, Route } from 'react-router-dom'
import joinURL from '../util/joinURL'
import RotationRouter from '../rotations/RotationRouter'
import AlertRouter from '../alerts/AlertRouter'
import ScheduleRouter from '../schedules/ScheduleRouter'
import PolicyRouter from '../escalation-policies/PolicyRouter'
import ServiceRouter from '../services/ServiceRouter'
import UserRouter from '../users/UserRouter'
import AdminRouter from '../admin/AdminRouter'
import WizardRouter from '../wizard/WizardRouter'
import IntegrationKeyAPI from '../documentation/components/IntegrationKeyAPI'

export const getPath = (p) => (Array.isArray(p.path) ? p.path[0] : p.path)

export function renderRoutes(routeConfig = []) {
  const routes = []

  routeConfig.forEach((cfg, idx) => {
    const _path = cfg.path
    const path = Array.isArray(_path) ? _path[0] : _path

    // redirect to remove trailing slashes
    routes.push(
      <Redirect
        key={`redir_${idx}`}
        strict
        exact
        from={path.replace(/\/?$/, '/')}
        to={path.replace(/\/$/, '')}
      />,
    )

    if (Array.isArray(_path)) {
      // add alias routes (for compatibility)
      _path.slice(1).forEach((p, pIdx) => {
        routes.push(
          <Redirect key={`alias_${idx}_${pIdx}`} exact from={p} to={path} />,
        )
        if (p !== '/') {
          // redirect nested paths (e.g. /on_call_schedules/foo to /schedules/foo)
          routes.push(
            <Redirect
              key={`alias_${idx}_${pIdx}_splat`}
              from={joinURL(p, '*')}
              to={joinURL(path, '*')}
            />,
          )
        }
      })
    }

    if (cfg.subRoutes && cfg.subRoutes.length) {
      routes.push(
        <Redirect
          key={`redir_sub_${idx}`}
          strict
          exact
          from={path.replace(/\/$/, '')}
          to={cfg.subRoutes[0].path}
        />,
      )
    }

    routes.push(
      <Route
        key={'route_' + idx}
        render={() => <cfg.component />}
        path={path}
        exact={cfg.exact}
      />,
    )
  })

  return routes
}

// used by new app and the toolbar title
export default [
  {
    title: 'Alerts',
    path: ['/alerts', '/'],
    component: AlertRouter,
  },
  {
    title: 'Rotations',
    path: '/rotations',
    component: RotationRouter,
  },
  {
    title: 'Schedules',
    path: ['/schedules', '/on_call_schedules'],
    component: ScheduleRouter,
  },
  {
    title: 'Escalation Policies',
    path: ['/escalation-policies', '/escalation_policies'],
    component: PolicyRouter,
  },
  {
    title: 'Services',
    path: '/services',
    component: ServiceRouter,
  },
  {
    title: 'Users',
    path: '/users',
    component: UserRouter,
  },
  {
    nav: false,
    title: 'Setup Wizard',
    path: '/wizard',
    component: WizardRouter,
  },
  {
    nav: false,
    title: 'Profile',
    path: '/profile',
    component: UserRouter,
  },
  {
    nav: false,
    title: 'Admin',
    path: '/admin',
    component: AdminRouter,
    subRoutes: [
      {
        title: 'Config',
        path: '/admin/config',
        component: AdminRouter,
      },
      {
        title: 'System Limits',
        path: '/admin/limits',
        component: AdminRouter,
      },
      {
        title: 'Toolbox',
        path: '/admin/toolbox',
        component: AdminRouter,
      },
    ],
  },
  {
    nav: false,
    title: 'Documentation',
    path: '/docs',
    component: IntegrationKeyAPI,
  },
]
