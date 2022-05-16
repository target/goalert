import * as React from 'react'
import Link, { LinkProps } from '@mui/material/Link'
import Typography from '@mui/material/Typography'
import Breadcrumbs from '@mui/material/Breadcrumbs'
import { ChevronRight } from '@mui/icons-material'
import { useQuery } from 'urql'
import { Link as RouterLink, Route, Routes, useParams } from 'react-router-dom'
import { applicationName as appName } from '../../env'
import { Theme } from '@mui/material'

const detailsMap: { [key: string]: string } = {
  alerts: 'alert',
  schedules: 'schedule',
  'escalation-policies': 'escalation Policy',
  rotations: 'rotation',
  users: 'user',
  services: 'service',
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
        textTransform: 'capitalize',
        color: getContrastColor,
      }}
    >
      {title.replace('-', ' ').replace('On Call', 'On-Call')}
    </Typography>
  )
}

// todo: handle profile
// todo: handle admin pages
// todo: fix lowercase + hyphens in tab title
// todo: fix escalation policies
function ToolbarBreadcrumbs(p: { isProfile?: boolean }): JSX.Element {
  const { sub, type = '', id } = useParams()
  const details = detailsMap[type ?? '']
  const detailsTitle = details + ' Details'

  document.title = `${applicationName || appName} - ${
    sub || (type ? detailsTitle : type)
  }`

  console.log('detailsTitle: ', detailsTitle)

  const [result] = useQuery({
    query: `query ($id: ID!) {
        data: ${details}(id: $id) {
          id
          name
        }
      }`,
    variables: { id },
    pause: !id,
  })

  console.log(result)

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
        {/* todo: remove hyphen for EPs */}
        {renderText(type, true)}
      </LinkRouter>
      {/* fix plural for Xs Details pages */}
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
      {id && sub && renderText(sub)}
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
