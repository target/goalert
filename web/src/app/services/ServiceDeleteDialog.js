import React from 'react'
import p from 'prop-types'

import { graphql2Client } from '../apollo'
import gql from 'graphql-tag'
import { Mutation } from 'react-apollo'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import Checkbox from '@material-ui/core/Checkbox'
import FormControlLabel from '@material-ui/core/FormControlLabel'
import FormControl from '@material-ui/core/FormControl'
import FormHelperText from '@material-ui/core/FormHelperText'
import { Redirect } from 'react-router-dom'
import Query from '../util/Query'

import FormDialog from '../dialogs/FormDialog'

class DeleteForm extends React.PureComponent {
  static propTypes = {
    epName: p.string.isRequired,
    error: p.string,
    value: p.bool,
    onChange: p.func.isRequired,
  }

  render() {
    return (
      <FormControl error={Boolean(this.props.error)} style={{ width: '100%' }}>
        <FormControlLabel
          control={
            <Checkbox
              checked={this.props.value}
              onChange={e => this.props.onChange(e.target.checked)}
              value='delete-escalation-policy'
            />
          }
          label={`Also delete escalation policy: ${this.props.epName}`}
        />
        <FormHelperText>{this.props.error}</FormHelperText>
      </FormControl>
    )
  }
}

const query = gql`
  query service($id: ID!) {
    service(id: $id) {
      id
      name
      escalationPolicyID
      escalationPolicy {
        id
        name
      }
    }
  }
`
const mutation = gql`
  mutation delete($input: [TargetInput!]!) {
    deleteAll(input: $input)
  }
`

export default class ServiceDeleteDialog extends React.PureComponent {
  static propTypes = {
    serviceID: p.string.isRequired,
    onClose: p.func,
  }

  state = {
    deleteEP: true,
  }

  renderQuery() {
    return (
      <Query
        noPoll
        client={graphql2Client}
        query={query}
        variables={{ id: this.props.serviceID }}
        render={({ data }) => this.renderMutation(data.service)}
      />
    )
  }

  renderMutation(svcData) {
    return (
      <Mutation client={graphql2Client} mutation={mutation}>
        {(commit, status) => this.renderDialog(svcData, commit, status)}
      </Mutation>
    )
  }

  renderDialog = (svcData, commit, mutStatus) => {
    const { loading, error, data } = mutStatus
    if (data && data.deleteAll) {
      return <Redirect push to={`/services`} />
    }

    return (
      <FormDialog
        title='Are you sure?'
        confirm
        subTitle={`This will delete the service: ${svcData.name}`}
        caption='Deleting a service will also delete all associated integration keys and alerts.'
        loading={loading}
        errors={nonFieldErrors(error)}
        onClose={this.props.onClose}
        onSubmit={() => {
          const input = [
            {
              type: 'service',
              id: this.props.serviceID,
            },
          ]
          if (this.state.deleteEP) {
            input.push({
              type: 'escalationPolicy',
              id: svcData.escalationPolicyID,
            })
          }
          return commit({
            variables: {
              input,
            },
          })
        }}
        form={
          <DeleteForm
            epName={svcData.escalationPolicy.name}
            error={
              fieldErrors(error).find(f => f.field === 'escalationPolicyID') &&
              'Escalation policy is currently in use.'
            }
            onChange={deleteEP => this.setState({ deleteEP })}
            value={this.state.deleteEP}
          />
        }
      />
    )
  }

  render() {
    return this.renderQuery()
  }
}
