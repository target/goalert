import * as React from 'react'
import Typography from '@mui/material/Typography'
import Breadcrumbs from '@mui/material/Breadcrumbs'
import { ChevronRight } from '@mui/icons-material'
import { useQuery } from 'urql'
import { useLocation } from 'wouter'
import { Theme } from '@mui/material'
import { startCase, camelCase } from 'lodash'
import { applicationName as appName } from '../../env'
import { routes } from '../AppRoutes'
import { useConfigValue } from '../../util/RequireConfig'
import AppLink from '../../util/AppLink'

// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-ignore type definition is broken for this file
import makeMatcher from 'wouter/matcher'
import { useIsWidthDown } from '../../util/useWidth'

const typeMap: { [key: string]: string } = {
  alerts: 'Alert',
  schedules: 'Schedule',
  'escalation-policies': 'Escalation Policy',
  rotations: 'Rotation',
  users: 'User',
  services: 'Service',
  'integration-keys': 'IntegrationKey',
}
const toTitleCase = (str: string): string =>
  startCase(str)
    .replace(/^Wizard/, 'Setup Wizard')
    .replace('On Call', 'On-Call')
    .replace('Docs', 'Documentation')
    .replace('Limits', 'System Limits')
    .replace('Admin ', 'Admin: ')
    .replace(/Config$/, 'Configuration')
    .replace('Api', 'API')

// todo: not needed once appbar is using same color prop for dark/light modes
const getContrastColor = (theme: Theme): string => {
  return theme.palette.getContrastText(
    theme.palette.mode === 'dark'
      ? theme.palette.background.paper
      : theme.palette.primary.main,
  )
}

const renderCrumb = (
  index: number,
  title: string,
  link?: string,
): JSX.Element => {
  const text = (
    <Typography
      data-cy={`breadcrumb-${index}`}
      noWrap
      key={index}
      component='h6'
      sx={{
        padding: '0 4px 0 4px',
        fontSize: '1.25rem',
        color: getContrastColor,
      }}
    >
      {title}
    </Typography>
  )

  if (!link) {
    return text
  }

  return (
    <AppLink
      key={link}
      to={link}
      underline='hover'
      color='inherit'
      sx={{
        '&:hover': {
          textDecoration: 'none',
        },
        '&:hover > h6': {
          cursor: 'pointer',
          backgroundColor: 'rgba(255, 255, 255, 0.2)',
          borderRadius: '6px',
          padding: '4px',
        },
      }}
    >
      {text}
    </AppLink>
  )
}

const matchPath = makeMatcher()

function useName(type = '', id = ''): string {
  const queryName = camelCase(typeMap[type] ?? 'skipping')
  const isUUID =
    /[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}/.test(id)
  // query for details page name if on a subpage
  const [result] = useQuery({
    query: `query ($id: ID!) {
        data: ${queryName}(id: $id) {
          id
          name
        }
      }`,
    variables: { id },
    pause: !type || !isUUID,
  })

  if (type && isUUID && result?.data?.data?.name) {
    return result.data.data.name
  }

  return toTitleCase(isUUID ? typeMap[type] ?? type : id)
}

function useBreadcrumbs(): [string, JSX.Element[] | JSX.Element] {
  const [path] = useLocation()

  let title = ''
  const crumbs: Array<JSX.Element> = []
  const parts = path.split('/')
  const name = useName(parts[1], parts[2])
  parts.slice(1).forEach((p, i) => {
    const part = decodeURIComponent(p)
    title = i === 1 ? name : toTitleCase(part)
    if (parts[1] === 'admin') {
      // admin doesn't have IDs to lookup
      // and instead just has fixed sub-page names
      title = toTitleCase(part)
    }
    crumbs.push(renderCrumb(i, title, parts.slice(0, i + 2).join('/')))
  })

  const isValidRoute = Object.keys(routes).some((pattern) => {
    const [match] = matchPath(pattern, path)
    return match
  })

  if (!isValidRoute) {
    const title = toTitleCase('page-not-found')
    return [title, renderCrumb(0, title)]
  }

  if (/^\d+$/.test(title)) {
    title = 'Alert ' + title
  }

  return [title, crumbs]
}

export default function ToolbarPageTitle(): JSX.Element {
  const [title, crumbs] = useBreadcrumbs()
  const [applicationName] = useConfigValue('General.ApplicationName')
  const isMobile = useIsWidthDown('md')

  React.useLayoutEffect(() => {
    document.title = `${applicationName || appName} - ${title}`
  }, [title, applicationName])

  return (
    <Breadcrumbs
      maxItems={isMobile ? 2 : undefined}
      separator={
        <ChevronRight
          sx={{
            color: getContrastColor,
          }}
        />
      }
    >
      {crumbs}
    </Breadcrumbs>
  )
}
