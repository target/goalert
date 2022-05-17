import * as React from 'react'
import Link from '@mui/material/Link'
import Typography from '@mui/material/Typography'
import Breadcrumbs from '@mui/material/Breadcrumbs'
import { ChevronRight } from '@mui/icons-material'
import { useQuery } from 'urql'
import { Link as RouterLink, Route, Routes, useParams } from 'react-router-dom'
import { Theme } from '@mui/material'
import { startCase, camelCase } from 'lodash'
import { applicationName as appName } from '../../env'

const typeMap: { [key: string]: string } = {
  alerts: 'Alert',
  schedules: 'Schedule',
  'escalation-policies': 'Escalation Policy',
  rotations: 'Rotation',
  users: 'User',
  services: 'Service',
}

// todo: not needed once appbar is using same color prop for dark/light modes
const getContrastColor = (theme: Theme): string => {
  return theme.palette.getContrastText(
    theme.palette.mode === 'dark'
      ? theme.palette.background.paper
      : theme.palette.primary.main,
  )
}

const renderText = (title: string, link?: string): JSX.Element => {
  const typography = (
    <Typography
      noWrap
      component='h1'
      sx={{
        padding: '0 4px 0 4px',
        fontSize: '1.25rem',
        color: getContrastColor,
      }}
    >
      {startCase(title.replace('-', ' ').replace('On Call', 'On-Call'))}
    </Typography>
  )

  if (link) {
    return (
      <Link
        component={RouterLink}
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
        {typography}
      </Link>
    )
  }

  return typography
}

function ToolbarBreadcrumbs(p: { type?: string }): JSX.Element {
  const typeFallback = p.type ?? ''
  const { sub, type = typeFallback, id } = useParams()

  const queryName = camelCase(typeMap[type]) ?? 'skipping'
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
      aria-label='breadcrumbs'
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

export default function ToolbarPageTitle(): JSX.Element {
  return (
    <Routes>
      {/* standard list/details/subpage route paths */}
      <Route path='/:type' element={<ToolbarBreadcrumbs />} />
      <Route path='/:type/:id' element={<ToolbarBreadcrumbs />} />
      <Route path='/:type/:id/:sub' element={<ToolbarBreadcrumbs />} />

      {/* everything else */}
      <Route
        path='/profile/:sub'
        element={<ToolbarBreadcrumbs type='profile' />}
      />
      <Route path='/admin/:sub' element={<ToolbarBreadcrumbs type='admin' />} />
      <Route path='/wizard' element={renderText('Setup Wizard')} />
      <Route path='/docs' element={renderText('Documentation')} />
    </Routes>
  )
}
