import React from 'react'
import p from 'prop-types'
import Typography from '@material-ui/core/Typography'
import { Switch, Route } from 'react-router-dom'
import { makeStyles } from '@material-ui/core/styles'
import { ChevronRight } from '@material-ui/icons'
import { gql, useQuery } from '@apollo/client'
import { startCase } from 'lodash'
import AppLink from '../../util/AppLink'
import { useIsWidthDown } from '../../util/useWidth'
import { useConfigValue } from '../../util/RequireConfig'
import { applicationName as appName } from '../../env'

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
    // flex: 1, // pushes toolbar actions to the right
    fontSize: '1.25rem',
  },
}))

const mapSingular = {
  Schedules: 'Schedule',
  'Escalation Policies': 'Escalation Policy',
  Rotations: 'Rotation',
  Users: 'User',
  Services: 'Service',
  Alerts: 'Alert',
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
  alerts: gql`
    query ($id: Int!) {
      data: alert(id: $id) {
        id
      }
    }
  `,
}

function NameLoader(props) {
  const { data } = useQuery(props.query, {
    variables: { id: props.id },
    skip: !props.id,
  })
  return data?.data?.name ?? props.id ?? props.fallback
}

NameLoader.propTypes = {
  fallback: p.string.isRequired,
  id: p.string,
  query: p.object.isRequired,
}

function ToolbarTitle() {
  const fullScreen = useIsWidthDown('md')
  const classes = useStyles()
  const [applicationName] = useConfigValue('General.ApplicationName')

  const renderTitle = (title) => {
    document.title = `${applicationName || appName} - ${title}`

    return (
      <Typography
        className={classes.backPage}
        color='inherit'
        noWrap
        component={AppLink}
        variant='h6'
        to='..'
        replace
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
    console.log('the Match is:', match)
    const sub = startCase(match.params.sub)

    if (fullScreen) {
      // mobile, only render current title
      return renderTitle(sub)
    }
    const query = queries[match.params.type]
    const queryTwo = queries.alerts

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
        <ChevronRight />
        {match.params.subID && (
          <Typography
            className={classes.title}
            color='inherit'
            noWrap
            variant='h6'
            to='..'
            replace
          >
            {query && (
              <NameLoader
                id={parseInt(match.params.subID)}
                query={queryTwo}
                fallback={detailsText(match)}
              />
            )}
          </Typography>
        )}
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
        path='/:type(services)/:id/:sub(alerts|integration-keys|heartbeat-monitors|labels)/:subID'
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
        path='/:type(schedules)/:id/:sub(assignments|on-call-notifications|escalation-policies|overrides|shifts)'
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
