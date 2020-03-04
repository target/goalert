import React from 'react'
import p from 'prop-types'
import Grid from '@material-ui/core/Grid'
import Card from '@material-ui/core/Card'
import CardContent from '@material-ui/core/CardContent'
import withStyles from '@material-ui/core/styles/withStyles'
import gql from 'graphql-tag'
import CreateFAB from '../lists/CreateFAB'
import FlatList from '../lists/FlatList'
import Query from '../util/Query'
import IconButton from '@material-ui/core/IconButton'
import { Trash } from '../icons'
import IntegrationKeyCreateDialog from './IntegrationKeyCreateDialog'
import IntegrationKeyDeleteDialog from './IntegrationKeyDeleteDialog'
import RequireConfig from '../util/RequireConfig'
import CopyText from '../util/CopyText'
import { AppLink } from '../util/AppLink'

const query = gql`
  query($serviceID: ID!) {
    service(id: $serviceID) {
      id # need to tie the result to the correct record
      integrationKeys {
        id
        type
        name
        href
      }
    }
  }
`

const styles = {
  copyIcon: {
    paddingRight: '0.25em',
    color: 'black',
  },
  keyLink: {
    display: 'flex',
    alignItems: 'center',
    width: 'fit-content',
  },
  spacing: {
    marginBottom: 96,
  },
}

const sortItems = (a, b) => {
  if (a.name.toLowerCase() < b.name.toLowerCase()) return -1
  if (a.name.toLowerCase() > b.name.toLowerCase()) return 1
  if (a.name < b.name) return -1
  if (a.name > b.name) return 1
  return 0
}

@withStyles(styles)
class IntegrationKeyDetails extends React.PureComponent {
  static propTypes = {
    href: p.string.isRequired,
    label: p.string.isRequired,
    type: p.string.isRequired,

    // provided by withStyles
    classes: p.object,
  }

  state = {
    showTooltip: false,
  }

  render() {
    let tooltip = (
      <CopyText title={'Copy ' + this.props.label} value={this.props.href} />
    )

    // if link is not properly present, do not display to copy
    if (this.props.type === 'email' && !this.props.href.startsWith('mailto:')) {
      tooltip = null
    }

    return (
      <React.Fragment>
        {tooltip}
        {this.props.type === 'email' && (
          <RequireConfig
            configID='Mailgun.Enable'
            else='Email integration keys are currently disabled.'
          />
        )}
      </React.Fragment>
    )
  }
}

@withStyles(styles)
export default class IntegrationKeyList extends React.PureComponent {
  static propTypes = {
    serviceID: p.string.isRequired,
  }

  state = {
    create: false,
    delete: null,
  }

  render() {
    return (
      <React.Fragment>
        <Grid item xs={12} className={this.props.classes.spacing}>
          <Card>
            <CardContent>{this.renderQuery()}</CardContent>
          </Card>
        </Grid>
        <CreateFAB
          onClick={() => this.setState({ create: true })}
          title='Create Integration Key'
        />
        {this.state.create && (
          <IntegrationKeyCreateDialog
            serviceID={this.props.serviceID}
            onClose={() => this.setState({ create: false })}
          />
        )}
        {this.state.delete && (
          <IntegrationKeyDeleteDialog
            integrationKeyID={this.state.delete}
            onClose={() => this.setState({ delete: null })}
          />
        )}
      </React.Fragment>
    )
  }

  renderQuery() {
    return (
      <Query
        query={query}
        variables={{ serviceID: this.props.serviceID }}
        render={({ data }) => this.renderList(data.service.integrationKeys)}
      />
    )
  }

  renderList(keys) {
    const typeLabels = {
      generic: 'Generic API Key',
      grafana: 'Grafana Webhook URL',
      site24x7: 'Site24x7 Webhook URL',
      email: 'Email Address',
    }

    const items = (keys || [])
      .slice()
      .sort(sortItems)
      .map(key => ({
        title: key.name,
        subText: (
          <IntegrationKeyDetails
            key={key.id}
            href={key.href}
            label={typeLabels[key.type]}
            type={key.type}
            classes={this.props.classes}
          />
        ),
        secondaryAction: (
          <IconButton onClick={() => this.setState({ delete: key.id })}>
            <Trash />
          </IconButton>
        ),
      }))

    return (
      <FlatList
        data-cy='int-keys'
        headerNote={
          <React.Fragment>
            API Documentation is available <AppLink to='/docs'>here</AppLink>.
          </React.Fragment>
        }
        emptyMessage='No integration keys exist for this service.'
        items={items}
      />
    )
  }
}
