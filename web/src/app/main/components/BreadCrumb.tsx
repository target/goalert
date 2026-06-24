import React from 'react'
import { Theme, Typography } from '@mui/material'
import AppLink from '../../util/AppLink'
import { camelCase, startCase } from 'lodash'
import { useQuery } from 'urql'

interface BreadCrumbProps {
  crumb: Routes | string
  urlParts: string[]
  index: number
  link?: string
}

type Routes =
  | 'alerts'
  | 'schedules'
  | 'rotations'
  | 'users'
  | 'escalation-policies'
  | 'services'
  | 'integration-keys'

// todo: not needed once appbar is using same color prop for dark/light modes
const getContrastColor = (theme: Theme): string => {
  return theme.palette.getContrastText(
    theme.palette.mode === 'dark'
      ? theme.palette.background.paper
      : theme.palette.primary.main,
  )
}

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

export function useName(crumb: string, index: number, parts: string[]): string {
  const isUUID =
    /[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}/.test(crumb)

  let type = typeMap[crumb]
  let queryName = 'skipping'
  if (isUUID) {
    type = parts[index - 1] // if uuid, grab type from previous crumb
    queryName = camelCase(typeMap[parts[index - 1]])
  }

  // query for details page name if on a subpage
  // skip if not a uuid - use type
  const [result] = useQuery({
    query: `query ($id: ID!) {
        data: ${queryName}(id: $id) {
          id
          name
        }
      }`,
    variables: { id: crumb },
    pause: !isUUID,
  })

  if (type && isUUID && result?.data?.data?.name) {
    return result.data.data.name
  }

  return toTitleCase(crumb)
}

export default function BreadCrumb(props: BreadCrumbProps): React.JSX.Element {
  const { index, crumb, link, urlParts } = props

  const name = useName(crumb, index, urlParts)

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
      {name}
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
