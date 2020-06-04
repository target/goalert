import React from 'react'
import { PropTypes as p } from 'prop-types'
import gql from 'graphql-tag'
import { Mutation } from 'react-apollo'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import PolicyStepForm from './PolicyStepForm'
import FormDialog from '../dialogs/FormDialog'
import { resetURLParams } from '../actions'
import { urlParamSelector } from '../selectors'
import { connect } from 'react-redux'

const mutation = gql`
  mutation($input: UpdateEscalationPolicyStepInput!) {
    updateEscalationPolicyStep(input: $input)
  }
`

@connect(
  (state) => ({
    errorMessage: urlParamSelector(state)('errorMessage'),
    errorTitle: urlParamSelector(state)('errorTitle'),
  }),
  (dispatch) => ({
    resetError: () => dispatch(resetURLParams('errorMessage', 'errorTitle')),
  }),
)
export default class PolicyStepEditDialog extends React.Component {
  static propTypes = {
    escalationPolicyID: p.string.isRequired,
    onClose: p.func.isRequired,
    step: p.shape({
      id: p.string.isRequired,
      // number from backend, string from textField
      delayMinutes: p.oneOfType([p.number, p.string]).isRequired,
      targets: p.arrayOf(
        p.shape({
          id: p.string.isRequired,
          name: p.string.isRequired,
          type: p.string.isRequired,
        }),
      ).isRequired,
    }),
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
        title='Edit Step'
        loading={loading}
        errors={nonFieldErrors(error)}
        maxWidth='sm'
        onClose={this.props.onClose}
        onSubmit={() =>
          commit({
            variables: {
              input: {
                id: this.props.step.id,
                delayMinutes:
                  (value && value.delayMinutes) || defaultValue.delayMinutes,
                targets: (value && value.targets) || defaultValue.targets,
              },
            },
          })
        }
        form={
          <PolicyStepForm
            errors={fieldErrs}
            disabled={loading}
            value={this.state.value || defaultValue}
            onChange={(value) => this.setState({ value })}
          />
        }
      />
    )
  }

  render() {
    const defaultValue = {
      targets: this.props.step.targets.map(({ id, type }) => ({ id, type })),
      delayMinutes: this.props.step.delayMinutes.toString(),
    }

    return (
      <Mutation mutation={mutation} onCompleted={this.props.onClose}>
        {(commit, status) => this.renderDialog(defaultValue, commit, status)}
      </Mutation>
    )
  }
}
