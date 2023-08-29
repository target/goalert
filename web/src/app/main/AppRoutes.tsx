import React, { JSXElementConstructor, useLayoutEffect } from 'react'
import { gql, useQuery } from 'urql'
import { Switch, Route, useLocation, RouteProps, useRoute } from 'wouter'
import AdminMessageLogsLayout from '../admin/admin-message-logs/AdminMessageLogsLayout'
import AdminAlertCounts from '../admin/admin-alert-counts/AdminAlertCounts'
import AdminConfig from '../admin/AdminConfig'
import AdminLimits from '../admin/AdminLimits'
import AdminToolbox from '../admin/AdminToolbox'
import AdminSwitchover from '../admin/switchover/AdminSwitchover'
import AlertsList from '../alerts/AlertsList'
import AlertDetailPage from '../alerts/pages/AlertDetailPage'
import Documentation from '../documentation/Documentation'
import { PageNotFound } from '../error-pages'
import PolicyDetails from '../escalation-policies/PolicyDetails'
import PolicyList from '../escalation-policies/PolicyList'
import PolicyServicesQuery from '../escalation-policies/PolicyServicesQuery'
import Spinner from '../loading/components/Spinner'
import RotationDetails from '../rotations/RotationDetails'
import RotationList from '../rotations/RotationList'
import ScheduleOnCallNotificationsList from '../schedules/on-call-notifications/ScheduleOnCallNotificationsList'
import ScheduleAssignedToList from '../schedules/ScheduleAssignedToList'
import ScheduleDetails from '../schedules/ScheduleDetails'
import ScheduleList from '../schedules/ScheduleList'
import ScheduleOverrideList from '../schedules/ScheduleOverrideList'
import ScheduleRuleList from '../schedules/ScheduleRuleList'
import ScheduleShiftList from '../schedules/ScheduleShiftList'
import AlertMetrics from '../services/AlertMetrics/AlertMetrics'
import HeartbeatMonitorList from '../services/HeartbeatMonitorList'
import IntegrationKeyList from '../services/IntegrationKeyList'
import ServiceAlerts from '../services/ServiceAlerts'
import ServiceDetails from '../services/ServiceDetails'
import ServiceLabelList from '../services/ServiceLabelList'
import ServiceList from '../services/ServiceList'
import UserCalendarSubscriptionList from '../users/UserCalendarSubscriptionList'
import UserDetails from '../users/UserDetails'
import UserList from '../users/UserList'
import UserOnCallAssignmentList from '../users/UserOnCallAssignmentList'
import UserSessionList from '../users/UserSessionList'
import { useSessionInfo } from '../util/RequireConfig'
import WizardRouter from '../wizard/WizardRouter'
import LocalDev from '../localdev/LocalDev'
import AdminSwitchoverGuide from '../admin/switchover/AdminSwitchoverGuide'

// ParamRoute will pass route parameters as props to the route's child.
function ParamRoute(props: RouteProps): JSX.Element {
  if (!props.path) {
    throw new Error('ParamRoute requires a path prop')
  }
  const [, params] = useRoute(props.path)
  const { component, children, ...rest } = props

  return (
    <Route {...rest}>
      {children
        ? React.Children.map(children as React.ReactChild, (child) =>
            React.cloneElement(child as React.ReactElement, params || {}),
          )
        : React.createElement(component as React.ComponentType, params)}
    </Route>
  )
}

const alertQuery = gql`
  query AlertQuery($id: Int!) {
    alert(id: $id) {
      id
      serviceID
    }
  }
`

// Allow any component to be used as a route.
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export const routes: Record<string, JSXElementConstructor<any>> = {
  '/': AlertsList,
  '/alerts': AlertsList,
  '/alerts/:alertID': AlertDetailPage,

  '/rotations': RotationList,
  '/rotations/:rotationID': RotationDetails,

  '/schedules': ScheduleList,
  '/schedules/:scheduleID': ScheduleDetails,
  '/schedules/:scheduleID/assignments': ScheduleRuleList,
  '/schedules/:scheduleID/escalation-policies': ScheduleAssignedToList,
  '/schedules/:scheduleID/overrides': ScheduleOverrideList,
  '/schedules/:scheduleID/shifts': ScheduleShiftList,
  '/schedules/:scheduleID/on-call-notifications':
    ScheduleOnCallNotificationsList,

  '/escalation-policies': PolicyList,
  '/escalation-policies/:policyID': PolicyDetails,
  '/escalation-policies/:policyID/services': PolicyServicesQuery,

  '/services': ServiceList,
  '/services/:serviceID': ServiceDetails,
  '/services/:serviceID/alerts': ServiceAlerts,
  '/services/:serviceID/alerts/:alertID': AlertDetailPage,
  '/services/:serviceID/heartbeat-monitors': HeartbeatMonitorList,
  '/services/:serviceID/integration-keys': IntegrationKeyList,
  '/services/:serviceID/labels': ServiceLabelList,
  '/services/:serviceID/alert-metrics': AlertMetrics,

  '/users': UserList,
  '/users/:userID': UserDetails,
  '/users/:userID/on-call-assignments': UserOnCallAssignmentList,
  '/users/:userID/schedule-calendar-subscriptions':
    UserCalendarSubscriptionList,
  '/users/:userID/sessions': UserSessionList,

  '/profile': Spinner, // should redirect once user ID loads
  '/profile/*': Spinner, // should redirect once user ID loads

  '/admin/config': AdminConfig,
  '/admin/limits': AdminLimits,
  '/admin/toolbox': AdminToolbox,
  '/admin/message-logs': AdminMessageLogsLayout,
  '/admin/alert-counts': AdminAlertCounts,
  '/admin/switchover': AdminSwitchover,
  '/admin/switchover/guide': AdminSwitchoverGuide,

  '/wizard': WizardRouter,
  '/docs': Documentation,
}

if (process.env.NODE_ENV !== 'production') {
  routes['/dev'] = LocalDev
}

export default function AppRoutes(): JSX.Element {
  const [path, setPath] = useLocation()
  const { userID } = useSessionInfo()

  const [, rootAlertInfo] = useRoute('/alerts/:alertID')
  const [, svcAlertInfo] = useRoute('/services/:serviceID/alerts/:alertID')
  const urlAlertID = rootAlertInfo?.alertID || svcAlertInfo?.alertID
  const urlServiceID = svcAlertInfo?.serviceID
  const [alertQ] = useQuery({
    query: alertQuery,
    pause: !urlAlertID,
    variables: { id: urlAlertID },
  })
  const alertServiceID = urlAlertID && alertQ.data?.alert?.serviceID

  useLayoutEffect(() => {
    if (path.endsWith('/') && path !== '/') {
      setPath(path.slice(0, -1) + location.search + location.hash, {
        replace: true,
      })
      return
    }

    const redirects: Record<string, string> = {
      '=/': '/alerts',
      '/on_call_schedules': '/schedules',
      '/escalation_policies': '/escalation-policies',
      '=/admin': '/admin/config',
    }
    if (userID) {
      redirects['/profile'] = `/users/${userID}`
    }
    if (alertServiceID && alertServiceID !== urlServiceID) {
      redirects[path] = `/services/${alertServiceID}/alerts/${urlAlertID}`
    }
    const redirect = (from: string, to: string): void => {
      setPath(to + path.slice(from.length) + location.search + location.hash, {
        replace: true,
      })
    }

    for (const [from, to] of Object.entries(redirects)) {
      if (from.startsWith('=') && path === from.slice(1)) {
        redirect(from.slice(1), to)
        return
      }

      if (!path.startsWith(from)) {
        continue
      }

      redirect(from, to)
      return
    }
  }, [path, userID, alertServiceID, urlAlertID, urlServiceID])

  return (
    <Switch>
      {
        Object.entries(routes).map(([path, component]) => (
          <ParamRoute key={path} path={path} component={component} />

          // not worth the type headache, we just want our routes
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
        )) as any
      }
      <Route component={PageNotFound} />
    </Switch>
  )
}
