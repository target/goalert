import React, { useEffect, useLayoutEffect } from 'react'
import p from 'prop-types'
import Typography from '@mui/material/Typography'
import makeStyles from '@mui/styles/makeStyles'
import { ChevronRight } from '@mui/icons-material'
import { gql, useQuery } from 'urql'
import { startCase } from 'lodash'
import AppLink from '../../util/AppLink'
import { useIsWidthDown } from '../../util/useWidth'
import { useConfigValue, useSessionInfo } from '../../util/RequireConfig'
import { applicationName as appName } from '../../env'
import { Route, Switch, useRoute } from 'wouter'
import _ from 'lodash'

const useStyles = makeStyles(() => ({
  backPage: {
    '&:hover': {
      cursor: 'pointer',
      backgroundColor: 'rgba(255, 255, 255, 0.2)',
      borderRadius: '6px',
      padding: '4px',
      textDecoration: 'none',
    },
    padding: '0 4px 0 4px',
  },
  div: {
    alignItems: 'center',
    display: 'flex',
    height: '100%',
    width: '100%',
  },
  title: {
    padding: '0 4px 0 4px',
    flex: 1, // pushes toolbar actions to the right
    fontSize: '1.25rem',
  },
}))

const mapSingular = {
  Schedules: 'Schedule',
  'Escalation Policies': 'Escalation Policy',
  Rotations: 'Rotation',
  Users: 'User',
  Services: 'Service',
}

const queries = {
  users: gql`
    query ($id: ID!) {
      data: user(id: $id) {
        id
        name
      }
    }
  `,
  services: gql`
    query ($id: ID!) {
      data: service(id: $id) {
        id
        name
      }
    }
  `,
  schedules: gql`
    query ($id: ID!) {
      data: schedule(id: $id) {
        id
        name
      }
    }
  `,
  'escalation-policies': gql`
    query ($id: ID!) {
      data: escalationPolicy(id: $id) {
        id
        name
      }
    }
  `,
}

function NameLoader(props) {
  const [{ data }] = useQuery({
    query: props.query,
    variables: { id: props.id },
    pause: !props.id,
  })
  return data?.data?.name ?? props.fallback
}

NameLoader.propTypes = {
  fallback: p.string.isRequired,
  id: p.string,
  query: p.object.isRequired,
}

const titleMap = {
  'On Call': 'On-Call',
  Docs: 'Documentation',
  Wizard: 'Setup Wizard',
}

function ToolbarTitle() {
  const fullScreen = useIsWidthDown('md')
  const classes = useStyles()
  const [applicationName] = useConfigValue('General.ApplicationName')

  const useTitle = (title) => {
    title = titleMap[title] ?? title
    useEffect(() => {
      document.title = `${applicationName || appName} - ${title}`
    }, [title, applicationName])

    return (
      <Typography
        className={classes.title}
        color='inherit'
        noWrap
        component='h1'
      >
        {title}
      </Typography>
    )
  }

  const detailsText = (type) => {
    const typeName = startCase(type).replace(/^Admin /, 'Admin: ')
    return (
      (mapSingular[typeName] || typeName) +
      (type !== 'profile' && !type.startsWith('admin') ? ' Details' : '')
    )
  }

  function mapInfo(info, userID) {
    if (info.type === 'users' && info.id === userID) {
      return {
        ...info,
        type: 'profile',
        id: null,
      }
    }

    return info
  }

  function SubPageTitle() {
    let [, { type, sub, id }] = useRoute('/:type/:id/:sub')
    const { userID } = useSessionInfo()
    const isProfile = type === 'user' && id === userID
    if (isProfile) type = 'profile'

    const title = useTitle(startCase(sub))
    if (fullScreen) {
      // mobile, only render current title
      return title
    }
    const query = queries[type]

    return (
      <div className={classes.div}>
        <Typography
          component={AppLink}
          className={classes.backPage}
          color='inherit'
          noWrap
          variant='h6'
          to='..'
        >
          {query && !isProfile ? (
            <NameLoader id={id} query={query} fallback={detailsText(type)} />
          ) : (
            detailsText(type)
          )}
        </Typography>
        <ChevronRight />
        {title}
      </div>
    )
  }

  function DetailsPageTitle() {
    let [, { type, id }] = useRoute('/:type/:id')
    const { userID } = useSessionInfo()

    if (type === 'users' && id === userID) {
      type = 'profile'
    }

    if (type === 'admin') {
      switch (id) {
        case 'config':
          type = 'admin configuration'
          break
        case 'limits':
          type = 'admin system limits'
          break
        default:
          type = 'admin ' + id
      }
    }

    return useTitle(detailsText(type))
  }

  function TopLevelTitle() {
    const [, { type }] = useRoute('/:type')
    return useTitle(startCase(type))
  }

  return (
    <Switch>
      <Route path='/:type' children={<TopLevelTitle />} />
      <Route path='/:type/:id' children={<DetailsPageTitle />} />
      <Route path='/:type/:id/:sub' children={<SubPageTitle />} />
    </Switch>
  )
}

export default ToolbarTitle
