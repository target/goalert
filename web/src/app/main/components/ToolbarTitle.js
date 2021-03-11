import React from 'react'
import p from 'prop-types'
import Typography from '@material-ui/core/Typography'
import { Switch, Route } from 'react-router-dom'
import { makeStyles } from '@material-ui/core/styles'
import { isWidthUp } from '@material-ui/core/withWidth'
import { ChevronRight } from '@material-ui/icons'
import { gql, useQuery } from '@apollo/client'
import { startCase } from 'lodash'
import AppLink from '../../util/AppLink'
import useWidth from '../../util/useWidth'

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
    query($id: ID!) {
      data: user(id: $id) {
        id
        name
      }
    }
  `,
  services: gql`
    query($id: ID!) {
      data: service(id: $id) {
        id
        name
      }
    }
  `,
  schedules: gql`
    query($id: ID!) {
      data: schedule(id: $id) {
        id
        name
      }
    }
  `,
  'escalation-policies': gql`
    query($id: ID!) {
      data: escalationPolicy(id: $id) {
        id
        name
      }
    }
  `,
}

function NameLoader(props) {
  const { data } = useQuery(props.query, {
    variables: { id: props.id },
    skip: !props.query || !props.id,
  })
  if (!data || !data.data) return props.fallback
  return data.data.name
}

NameLoader.propTypes = {
  fallback: p.string.isRequired,
  id: p.string,
  query: p.object,
}

function ToolbarTitle() {
  const width = useWidth()
  const classes = useStyles()

  const renderTitle = (title) => {
    document.title = `GoAlert - ${title}`

    return (
      <Typography
        className={classes.title}
        color='inherit'
        noWrap
        component='h1'
      >
        {title.replace('On Call', 'On-Call')}
      </Typography>
    )
  }

  const detailsText = (match) => {
    const typeName = startCase(match.params.type)
    return (
      (mapSingular[typeName] || typeName) +
      (match.params.type !== 'profile' ? ' Details' : '')
    )
  }

  const renderSubPageTitle = ({ match }) => {
    const sub = startCase(match.params.sub)

    if (!isWidthUp('md', width)) {
      // mobile, only render current title
      return renderTitle(sub)
    }
    const query = queries[match.params.type]

    return (
      <div className={classes.div}>
        <Typography
          component={AppLink}
          className={classes.backPage}
          color='inherit'
          noWrap
          variant='h6'
          to='..'
          replace
        >
          {query ? (
            <NameLoader
              id={match.params.id}
              query={query}
              fallback={detailsText(match)}
            />
          ) : (
            detailsText(match)
          )}
        </Typography>
        <ChevronRight />
        {renderTitle(sub)}
      </div>
    )
  }

  const renderDetailsPageTitle = ({ match }) => {
    return renderTitle(detailsText(match))
  }

  const renderTopLevelTitle = ({ match }) => {
    return renderTitle(startCase(match.params.type))
  }

  return (
    <Switch>
      <Route
        path='/:type(escalation-policies)/:id/:sub(services)'
        render={renderSubPageTitle}
      />
      <Route
        path='/:type(services)/:id/:sub(alerts|integration-keys|heartbeat-monitors|labels)'
        render={renderSubPageTitle}
      />
      <Route
        path='/:type(users)/:id/:sub(on-call-assignments|schedule-calendar-subscriptions|sessions)'
        render={renderSubPageTitle}
      />
      <Route
        path='/:type(profile)/:sub(on-call-assignments|schedule-calendar-subscriptions|sessions)'
        render={renderSubPageTitle}
      />
      <Route
        path='/:type(schedules)/:id/:sub(assignments|escalation-policies|overrides|shifts)'
        render={renderSubPageTitle}
      />
      <Route
        path='/:type(alerts|rotations|schedules|escalation-policies|services|users)/:id'
        render={renderDetailsPageTitle}
      />
      <Route
        path='/:type(alerts|rotations|schedules|escalation-policies|services|users|profile)'
        render={renderTopLevelTitle}
      />
      <Route path='/wizard' render={() => renderTitle('Setup Wizard')} />
      <Route path='/admin' render={() => renderTitle('Admin Page')} />
      <Route path='/docs' render={() => renderTitle('Documentation')} />
    </Switch>
  )
}

export default ToolbarTitle
