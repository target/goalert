import React from 'react'
import { PropTypes as p } from 'prop-types'
import { graphql2Client } from '../apollo'
import gql from 'graphql-tag'
import { Mutation } from 'react-apollo'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import PolicyStepForm from './PolicyStepForm'
import FormDialog from '../dialogs/FormDialog'
import { resetURLParams } from '../actions'
import { urlParamSelector } from '../selectors'
import { connect } from 'react-redux'

const mutation = gql`
  mutation($input: CreateEscalationPolicyStepInput!) {
    createEscalationPolicyStep(input: $input) {
      id
      delayMinutes
      targets {
        id
        name
        type
      }
    }
  }
`

const refetchQuery = gql`
  query($id: ID!) {
    escalationPolicy(id: $id) {
      id
      steps {
        id
        delayMinutes
        targets {
          id
          name
          type
        }
      }
    }
  }
`

@connect(
  state => ({
    errorMessage: urlParamSelector(state)('errorMessage'),
    errorTitle: urlParamSelector(state)('errorTitle'),
  }),
  dispatch => ({
    resetError: () => dispatch(resetURLParams('errorMessage', 'errorTitle')),
  }),
)
export default class PolicyStepCreateDialog extends React.Component {
  static propTypes = {
    escalationPolicyID: p.string.isRequired,
    onClose: p.func.isRequired,
  }

  state = {
    value: null,
    errors: [],
  }

  renderDialog(defaultValue, commit, status) {
    const { errorMessage, errorTitle } = this.props
    const { value } = this.state
    const { loading, error } = status
    const fieldErrs = fieldErrors(error)

    // don't render dialog if slack redirect returns with an error
    if (Boolean(errorMessage) || Boolean(errorTitle)) {
      return null
    }

    return (
      <FormDialog
        title='Create Step'
        loading={loading}
        errors={nonFieldErrors(error)}
        maxWidth='sm'
        onClose={this.props.onClose}
        onSubmit={() => {
          return commit({
            variables: {
              input: {
                escalationPolicyID: this.props.escalationPolicyID,
                delayMinutes: parseInt(
                  (value && value.delayMinutes) || defaultValue.delayMinutes,
                ),
                targets: (value && value.targets) || defaultValue.targets,
              },
            },
          }).then(() => this.props.onClose())
        }}
        form={
          <PolicyStepForm
            errors={fieldErrs}
            disabled={loading}
            value={this.state.value || defaultValue}
            onChange={value => this.setState({ value })}
          />
        }
      />
    )
  }

  render() {
    const defaultValue = {
      targets: [],
      delayMinutes: '15',
    }

    return (
      <Mutation
        client={graphql2Client}
        mutation={mutation}
        awaitRefetchQueries
        refetchQueries={() => [
          {
            query: refetchQuery,
            variables: {
              id: this.props.escalationPolicyID,
            },
          },
        ]}
      >
        {(commit, status) => this.renderDialog(defaultValue, commit, status)}
      </Mutation>
    )
  }
}
