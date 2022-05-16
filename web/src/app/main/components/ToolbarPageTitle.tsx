import * as React from 'react'
import Link, { LinkProps } from '@mui/material/Link'
import Typography from '@mui/material/Typography'
import Breadcrumbs from '@mui/material/Breadcrumbs'
import { ChevronRight } from '@mui/icons-material'
import { useQuery } from 'urql'
import { Link as RouterLink, Route, Routes, useParams } from 'react-router-dom'
import { Theme } from '@mui/material'
import { startCase } from 'lodash'
import { applicationName as appName } from '../../env'

const detailsMap: { [key: string]: string } = {
  alerts: 'Alert',
  schedules: 'Schedule',
  'escalation-policies': 'Escalation-Policy',
  rotations: 'Rotation',
  users: 'User',
  services: 'Service',
}

interface LinkRouterProps extends LinkProps {
  to: string
  replace?: boolean
}
const LinkRouter = (props: LinkRouterProps): JSX.Element => (
  <Link {...props} component={RouterLink as any} />
)

const getContrastColor = (theme: Theme): string => {
  return theme.palette.getContrastText(
    theme.palette.mode === 'dark'
      ? theme.palette.background.paper
      : theme.palette.primary.main,
  )
}

const renderText = (title: string, asLink?: boolean): JSX.Element => {
  let linkSx = {}
  if (asLink) {
    linkSx = {
      '&:hover': {
        cursor: 'pointer',
        backgroundColor: 'rgba(255, 255, 255, 0.2)',
        borderRadius: '6px',
        padding: '4px',
        textDecoration: 'none',
      },
    }
  }

  return (
    <Typography
      noWrap
      component='h1'
      sx={{
        ...linkSx,
        padding: '0 4px 0 4px',
        fontSize: '1.25rem',
        color: getContrastColor,
      }}
    >
      {title.replace('-', ' ').replace('On Call', 'On-Call')}
    </Typography>
  )
}

// todo: fix lowercase + hyphens in tab title
function ToolbarBreadcrumbs(p: { isProfile?: boolean }): JSX.Element {
  const { sub: _sub, type: _type, id } = useParams()
  const sub = startCase(_sub)
  const type = p.isProfile ? 'profile' : _type ?? '' // no type if on profile

  const details = detailsMap[type ?? ''] ?? startCase(type)
  const detailsTitle = details + ' Details'

  document.title = `${applicationName || appName} - ${
    sub || (type ? detailsTitle : type)
  }`

  const [result] = useQuery({
    pause: !id || !details,
    query: `query ($id: ID!) {
        data: ${details.replace('-', '')}(id: $id) {
          id
          name
        }
      }`,
    variables: { id },
  })

  return (
    <Breadcrumbs
      aria-label='breadcrumb'
      separator={
        <ChevronRight
          sx={{
            color: getContrastColor,
          }}
        />
      }
    >
      <LinkRouter
        underline='hover'
        color='inherit'
        to={'/' + type}
        key={'/' + type}
        sx={{
          textTransform: 'capitalize',
          '&:hover': { textDecoration: 'none' },
        }}
      >
        {renderText(type, true)}
      </LinkRouter>
      {id && type && !sub && renderText(detailsTitle)}
      {id && type && sub && (
        <LinkRouter
          key={'/' + type + '/' + id}
          to={'/' + type + '/' + id}
          underline='hover'
          color='inherit'
          sx={{
            textTransform: 'capitalize',
            '&:hover': { textDecoration: 'none' },
          }}
        >
          {renderText(result?.data?.data?.name ?? detailsTitle, true)}
        </LinkRouter>
      )}
      {sub && renderText(sub)}
    </Breadcrumbs>
  )
}

export default function ToolbarPageTitle(): JSX.Element {
  return (
    <Routes>
      <Route path='/:type' element={<ToolbarBreadcrumbs />} />
      <Route path='/:type/:id' element={<ToolbarBreadcrumbs />} />
      <Route path='/:type/:id/:sub' element={<ToolbarBreadcrumbs />} />

      <Route path='/profile/:sub' element={<ToolbarBreadcrumbs isProfile />} />
      <Route path='/wizard' element={renderText('Setup Wizard')} />
      <Route path='/admin' element={renderText('Admin Page')} />
      <Route path='/docs' element={renderText('Documentation')} />
    </Routes>
  )
}
