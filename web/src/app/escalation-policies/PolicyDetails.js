import React, { PureComponent } from 'react'
import p from 'prop-types'
import gql from 'graphql-tag'
import PageActions from '../util/PageActions'
import Query from '../util/Query'
import PolicyStepsQuery from './PolicyStepsQuery'
import OtherActions from '../util/OtherActions'
import PolicyDeleteDialog from './PolicyDeleteDialog'
import CreateFAB from '../lists/CreateFAB'
import PolicyStepCreateDialog from './PolicyStepCreateDialog'
import DetailsPage from '../details/DetailsPage'
import PolicyEditDialog from './PolicyEditDialog'
import { setURLParam } from '../actions/main'
import { connect } from 'react-redux'
import { urlParamSelector } from '../selectors'
import { resetURLParams } from '../actions'

const query = gql`
  query($id: ID!) {
    escalationPolicy(id: $id) {
      id
      name
      description
    }
  }
`

@connect(
  state => ({
    createStep: urlParamSelector(state)('createStep'),
  }),
  dispatch => ({
    setCreateStep: value => dispatch(setURLParam('createStep', value, false)),
    resetCreateStep: () => dispatch(resetURLParams('createStep')),
  }),
)
export default class PolicyDetails extends PureComponent {
  static propTypes = {
    escalationPolicyID: p.string.isRequired,
  }

  state = {
    delete: false,
    edit: false,
  }

  renderData = ({ data }) => {
    return (
      <React.Fragment>
        <PageActions>
          <OtherActions
            actions={[
              {
                label: 'Edit Escalation Policy',
                onClick: () => this.setState({ edit: true }),
              },
              {
                label: 'Delete Escalation Policy',
                onClick: () => this.setState({ delete: true }),
              },
            ]}
          />
        </PageActions>
        <DetailsPage
          title={data.escalationPolicy.name}
          details={data.escalationPolicy.description}
          links={[
            {
              label: 'Services',
              url: 'services',
            },
          ]}
          pageFooter={
            <PolicyStepsQuery escalationPolicyID={data.escalationPolicy.id} />
          }
        />
        <CreateFAB onClick={() => this.props.setCreateStep(true)} />
        {this.props.createStep && (
          <PolicyStepCreateDialog
            escalationPolicyID={data.escalationPolicy.id}
            onClose={this.props.resetCreateStep}
          />
        )}
        {this.state.edit && (
          <PolicyEditDialog
            escalationPolicyID={data.escalationPolicy.id}
            onClose={() => this.setState({ edit: false })}
          />
        )}
        {this.state.delete && (
          <PolicyDeleteDialog
            escalationPolicyID={data.escalationPolicy.id}
            onClose={() => this.setState({ delete: false })}
          />
        )}
      </React.Fragment>
    )
  }

  render() {
    return (
      <Query
        query={query}
        render={this.renderData}
        variables={{ id: this.props.escalationPolicyID }}
      />
    )
  }
}
