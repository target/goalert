import React from 'react'
import Query from '../util/Query'
import Button from '@material-ui/core/Button'
import Card from '@material-ui/core/Card'
import Grid from '@material-ui/core/Grid'
import Typography from '@material-ui/core/Typography'
import gql from 'graphql-tag'
import { chain, startCase, isEmpty } from 'lodash-es'
import AdminConfigSection from './AdminConfigSection'

import withStyles from '@material-ui/core/styles/withStyles'
import AdminConfirmDialog from './AdminConfirmDialog'
import PageActions from '../util/PageActions'
import { Form } from '../forms'

const query = gql`
  query getConfig {
    config(all: true) {
      id
      description
      password
      type
      value
    }
  }
`

const styles = theme => ({
  gridContainer: {
    [theme.breakpoints.up('md')]: {
      justifyContent: 'center',
    },
  },
  gridItem: {
    [theme.breakpoints.up('md')]: {
      maxWidth: '65%',
    },
  },
  groupTitle: {
    fontSize: '1.1rem',
  },
  saveDisabled: {
    color: 'rgba(255, 255, 255, 0.5)',
  },
})

@withStyles(styles)
export default class AdminConfig extends React.PureComponent {
  state = {
    tab: 0,
    confirm: false,
    value: {},
  }

  updateValue = (id, value) => {
    const newVal = { ...this.state.value }

    if (value === null) {
      delete newVal[id]
    } else {
      newVal[id] = value
    }

    this.setState({ value: newVal })
  }

  render() {
    return (
      <Query
        query={query}
        render={({ data }) => this.renderTabs(data.config)}
      />
    )
  }

  renderTabs(configValues) {
    const groups = chain(configValues)
      .map(f => f.id.split('.')[0])
      .uniq()
      .value()

    return (
      <React.Fragment>
        <Grid
          container
          spacing={2}
          className={this.props.classes.gridContainer}
        >
          {groups.map((groupID, index) => (
            <Grid
              key={index}
              container // contains title above card/card itself
              item // for each admin config section
              xs={12}
              className={this.props.classes.gridItem}
            >
              <Grid item xs={12}>
                <Typography
                  component='h2'
                  variant='subtitle1'
                  color='textSecondary'
                  classes={{
                    subtitle1: this.props.classes.groupTitle,
                  }}
                >
                  {startCase(groupID).replace('Git Hub', 'GitHub')}
                </Typography>
              </Grid>
              <Grid item xs={12}>
                <Form>
                  <Card>
                    <AdminConfigSection
                      value={this.state.value}
                      onChange={(id, value) => this.updateValue(id, value)}
                      fields={configValues
                        .filter(f => f.id.split('.')[0] === groups[index])
                        .map(f => ({
                          id: f.id,
                          label: chain(f.id.split('.'))
                            .last()
                            .startCase()
                            .value()
                            .replace(/R Ls\b/, 'RLs'), // fix usages of `URLs`
                          description: f.description,
                          password: f.password,
                          type: f.type,
                          value: f.value,
                        }))}
                    />
                  </Card>
                </Form>
              </Grid>
            </Grid>
          ))}
        </Grid>
        <PageActions>
          <Button
            color='inherit'
            data-cy='reset'
            disabled={isEmpty(this.state.value)}
            onClick={() => this.setState({ value: {} })}
            classes={{
              label: isEmpty(this.state.value)
                ? this.props.classes.saveDisabled
                : null,
            }}
          >
            Reset
          </Button>
          <Button
            color='inherit'
            data-cy='save'
            disabled={isEmpty(this.state.value)}
            onClick={() => this.setState({ confirm: true })}
            classes={{
              label: isEmpty(this.state.value)
                ? this.props.classes.saveDisabled
                : null,
            }}
          >
            Save
          </Button>
        </PageActions>
        {this.state.confirm && (
          <AdminConfirmDialog
            configValues={configValues}
            fieldValues={this.state.value}
            onClose={() => this.setState({ confirm: false })}
            onComplete={() => this.setState({ confirm: false, value: {} })}
          />
        )}
      </React.Fragment>
    )
  }
}
