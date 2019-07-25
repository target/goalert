import React from 'react'
import p from 'prop-types'
import Typography from '@material-ui/core/Typography'
import { Switch, Route, Link } from 'react-router-dom'
import withWidth, { isWidthUp } from '@material-ui/core/withWidth'
import { ChevronRight } from '@material-ui/icons'
import withStyles from '@material-ui/core/styles/withStyles'
import gql from 'graphql-tag'
import { Query } from 'react-apollo'
import { startCase } from 'lodash-es'
import { graphql2Client } from '../../apollo'
import { connect } from 'react-redux'
import { absURLSelector } from '../../selectors/url'

const styles = {
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
}

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

class NameLoader extends React.PureComponent {
  static propTypes = {
    fallback: p.string.isRequired,
    id: p.string,
    query: p.object,
  }

  render() {
    if (!this.props.query || !this.props.id) return this.props.fallback
    return (
      <Query
        query={this.props.query}
        variables={{ id: this.props.id }}
        client={graphql2Client}
      >
        {({ data }) => {
          if (!data || !data.data) {
            return this.props.fallback
          }

          return data.data.name
        }}
      </Query>
    )
  }
}

const mapStateToProps = state => {
  return {
    absURL: absURLSelector(state),
  }
}

@withWidth()
@withStyles(styles)
@connect(mapStateToProps)
export default class ToolbarTitle extends React.Component {
  renderTitle = title => {
    document.title = `GoAlert - ${title}`

    return (
      <Typography
        className={this.props.classes.title}
        color='inherit'
        noWrap
        component='h1'
      >
        {title.replace('On Call', 'On-Call')}
      </Typography>
    )
  }

  renderSubPageTitle = ({ match }) => {
    const sub = startCase(match.params.sub)

    if (!isWidthUp('md', this.props.width)) {
      // mobile, only render current title
      return this.renderTitle(sub)
    }
    const query = queries[match.params.type]

    return (
      <div className={this.props.classes.div}>
        <Typography
          component={Link}
          className={this.props.classes.backPage}
          color='inherit'
          noWrap
          variant='h6'
          to={this.props.absURL('..')}
          replace
        >
          <NameLoader
            id={match.params.id}
            query={query}
            fallback={this.detailsText(match)}
          />
        </Typography>
        <ChevronRight />
        {this.renderTitle(sub)}
      </div>
    )
  }

  detailsText = match => {
    const typeName = startCase(match.params.type)
    return (
      (mapSingular[typeName] || typeName) +
      (match.params.type !== 'profile' ? ' Details' : '')
    )
  }

  renderDetailsPageTitle = ({ match }) => {
    return this.renderTitle(this.detailsText(match))
  }

  renderTopLevelTitle = ({ match }) => {
    return this.renderTitle(startCase(match.params.type))
  }

  render() {
    return (
      <Switch>
        <Route
          path='/:type(escalation-policies)/:id/:sub(services)'
          render={this.renderSubPageTitle}
        />
        <Route
          path='/:type(services)/:id/:sub(alerts|integration-keys|heartbeat-monitors|labels)'
          render={this.renderSubPageTitle}
        />
        <Route
          path='/:type(users)/:id/:sub(on-call-assignments)'
          render={this.renderSubPageTitle}
        />
        <Route
          path='/:type(profile)/:sub(on-call-assignments)'
          render={this.renderSubPageTitle}
        />
        <Route
          path='/:type(schedules)/:id/:sub(assignments|escalation-policies|overrides|shifts)'
          render={this.renderSubPageTitle}
        />
        <Route
          path='/:type(alerts|rotations|schedules|escalation-policies|services|users)/:id'
          render={this.renderDetailsPageTitle}
        />
        <Route
          path='/:type(alerts|rotations|schedules|escalation-policies|services|users|profile)'
          render={this.renderTopLevelTitle}
        />
        <Route path='/wizard' render={() => this.renderTitle('Setup Wizard')} />
        <Route path='/admin' render={() => this.renderTitle('Admin Page')} />
        <Route path='/docs' render={() => this.renderTitle('Documentation')} />
      </Switch>
    )
  }
}
