import React from 'react'
import Query from '../util/Query'
import Button from '@material-ui/core/Button'
import Card from '@material-ui/core/Card'
import Grid from '@material-ui/core/Grid'
import gql from 'graphql-tag'
import { chain, isEmpty } from 'lodash-es'

import withStyles from '@material-ui/core/styles/withStyles'
import AdminDialog from './AdminDialog'
import PageActions from '../util/PageActions'
import { Form } from '../forms'
import AdminSection from './AdminSection'

const query = gql`
  query getLimits {
    systemLimits {
      id
      description
      value
    }
  }
`
const mutation = gql`
  mutation($input: [SystemLimitInput!]!) {
    setSystemLimits(input: $input)
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
export default class AdminLimits extends React.PureComponent {
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
        render={({ data }) => this.renderForm(data.systemLimits)}
      />
    )
  }

  renderForm(limitValues) {
    return (
      <React.Fragment>
        <Grid
          container
          spacing={2}
          className={this.props.classes.gridContainer}
        >
          <Grid item xs={12} className={this.props.classes.gridItem}>
            <Grid item xs={12}>
              <Form>
                <Card>
                  <AdminSection
                    value={this.state.value}
                    onChange={(id, value) => this.updateValue(id, value)}
                    fields={limitValues.map(f => ({
                      id: f.id,
                      type: 'integer',
                      description: f.description,
                      value: f.value.toString(),
                      label: chain(f.id.replace(/([a-z])([A-Z])/g, '$1 $2'))
                        .startCase()
                        .value(),
                      password: false,
                    }))}
                  />
                </Card>
              </Form>
            </Grid>
          </Grid>
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
          <AdminDialog
            mutation={mutation}
            values={limitValues}
            fieldValues={this.state.value}
            onClose={() => this.setState({ confirm: false })}
            onComplete={() => this.setState({ confirm: false, value: {} })}
          />
        )}
      </React.Fragment>
    )
  }
}
