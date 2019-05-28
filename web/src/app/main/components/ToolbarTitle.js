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
  },
}

const mapSingular = {
  Schedules: 'Schedule',
  'Escalation Policies': 'Escalation Policy',
  Rotations: 'Rotation',
  Users: 'User',
  Services: 'Service',
}

const nameQuery = typeName => gql`
  query($id: ID!) {
    data: ${typeName}(id: $id) {
      id
      name
    }
  }
`

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
    return (
      <Typography
        className={this.props.classes.title}
        color='inherit'
        noWrap
        variant='h6'
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

    let query
    switch (match.params.type) {
      case 'users':
        query = nameQuery('user')
        break
      case 'services':
        query = nameQuery('service')
        break
      case 'schedules':
        query = nameQuery('schedule')
        break
      case 'escalation-policies':
        query = nameQuery('escalationPolicy')
        break
    }

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
          path='/:type(services)/:id/:sub(alerts|integration-keys|labels)'
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
