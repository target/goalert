import * as React from 'react'
import Typography from '@mui/material/Typography'
import Breadcrumbs from '@mui/material/Breadcrumbs'
import { ChevronRight } from '@mui/icons-material'
import { useQuery } from 'urql'
import { useLocation, useRoute } from 'wouter'
import { Theme } from '@mui/material'
import { startCase, camelCase } from 'lodash'
import { applicationName as appName } from '../../env'
import { routes } from '../AppRoutes'
import makeMatcher from 'wouter/matcher'
import { useConfigValue } from '../../util/RequireConfig'
import AppLink from '../../util/AppLink'

const typeMap: { [key: string]: string } = {
  alerts: 'Alert',
  schedules: 'Schedule',
  'escalation-policies': 'Escalation Policy',
  rotations: 'Rotation',
  users: 'User',
  services: 'Service',
}
const toTitleCase = (str: string) =>
  startCase(str)
    .replace(/^Wizard/, 'Setup Wizard')
    .replace('On Call', 'On-Call')
    .replace('Docs', 'Documentation')
    .replace('Limits', 'System Limits')
    .replace('Admin ', 'Admin: ')
    .replace(/Config$/, 'Configuration')

// todo: not needed once appbar is using same color prop for dark/light modes
const getContrastColor = (theme: Theme): string => {
  return theme.palette.getContrastText(
    theme.palette.mode === 'dark'
      ? theme.palette.background.paper
      : theme.palette.primary.main,
  )
}

const renderCrumb = (title: string, link?: string): JSX.Element => {
  const text = (
    <Typography
      data-cy={title}
      noWrap
      key={title}
      component='h1'
      sx={{
        padding: '0 4px 0 4px',
        fontSize: '1.25rem',
        color: getContrastColor,
      }}
    >
      {toTitleCase(title)}
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
        '&:hover > h1': {
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

function ToolbarBreadcrumbs(p: { type?: string }): JSX.Element {
  const typeFallback = p.type ?? ''
  const { sub, type = typeFallback, id } = useParams()

  const queryName = camelCase(typeMap[type] ?? 'skipping')
  const detailsTitle = typeMap[type] + ' Details'

  document.title = `${applicationName || appName} - ${
    sub || (type ? detailsTitle : type)
  }`

  // query for details page name if on a subpage
  const [result] = useQuery({
    pause: !sub,
    query: `query ($id: ID!) {
        data: ${queryName}(id: $id) {
          id
          name
        }
      }`,
    variables: { id },
  })

  return (
    <Breadcrumbs
      separator={
        <ChevronRight
          sx={{
            color: getContrastColor,
          }}
        />
      }
    >
      {renderText(type, sub || id ? '/' + type : '')}
      {id && type && !sub && renderText(detailsTitle)}
      {id &&
        type &&
        sub &&
        renderText(
          result?.data?.data?.name ?? detailsTitle,
          '/' + type + '/' + id,
        )}
      {sub && renderText(sub)}
    </Breadcrumbs>
  )
}

const matchPath = makeMatcher()

function useName(type: string = '', id: string = '') {
  const queryName = camelCase(typeMap[type] ?? 'skipping')
  // query for details page name if on a subpage
  const [result] = useQuery({
    query: `query ($id: ID!) {
        data: ${queryName}(id: $id) {
          id
          name
        }
      }`,
    variables: { id },
    pause: !type || !id || type === 'admin',
  })

  if (result?.data?.data?.name) {
    return result.data.data.name
  }

  return typeMap[type] ?? type
}

function useBreadcrumbs() {
  const [, info] = useRoute('/:type?/:id?/:subType?')
  const { type, id, subType } = info || {}
  const [path] = useLocation()
  const isValidRoute = Object.keys(routes).some((pattern) => {
    const [match] = matchPath(pattern, path)
    return match
  })
  const name = useName(type, id)

  let title
  const crumbs: Array<JSX.Element> = []

  const push = (text: string, link?: string) => {
    crumbs.push(renderCrumb(text, link))
    title = toTitleCase(text)
  }

  if (!isValidRoute) {
    push('page-not-found')
  } else if (type === 'admin') {
    push('admin ' + id)
  } else if (type && id && subType) {
    push(name, '/' + type + '/' + id)
    push(subType)
  } else if (type && id) {
    push((typeMap[type] ?? type) + ' details')
  } else if (type) {
    push(type)
  }

  return [title, crumbs]
}

export default function ToolbarPageTitle(): JSX.Element {
  const [title, crumbs] = useBreadcrumbs()
  const [applicationName] = useConfigValue('General.ApplicationName')

  React.useLayoutEffect(() => {
    document.title = `${applicationName || appName} - ${title}`
  }, [title, applicationName])

  return (
    <Breadcrumbs
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
