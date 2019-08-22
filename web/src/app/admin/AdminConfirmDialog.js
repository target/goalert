import React from 'react'
import p from 'prop-types'
import List from '@material-ui/core/List'
import ListItem from '@material-ui/core/ListItem'
import ListItemText from '@material-ui/core/ListItemText'
import Typography from '@material-ui/core/Typography'
import gql from 'graphql-tag'
import { omit } from 'lodash-es'
import FormDialog from '../dialogs/FormDialog'
import { Mutation } from 'react-apollo'
import { graphql2Client } from '../apollo'
import { nonFieldErrors, fieldErrors } from '../util/errutil'
import Diff from '../util/Diff'

const mutation = gql`
  mutation($input: [ConfigValueInput!]) {
    setConfig(input: $input)
  }
`

export default class AdminConfirmDialog extends React.PureComponent {
  static propTypes = {
    configValues: p.array.isRequired,
    fieldValues: p.object.isRequired,
    onClose: p.func.isRequired,
    onComplete: p.func.isRequired,
  }

  render() {
    return (
      <Mutation
        client={graphql2Client}
        mutation={mutation}
        onCompleted={this.props.onComplete}
        awaitRefetchQueries
        refetchQueries={['getConfig']}
      >
        {(commit, status) => this.renderConfirm(commit, status)}
      </Mutation>
    )
  }

  renderConfirm(commit, { error }) {
    const changeKeys = Object.keys(this.props.fieldValues)
    const changes = this.props.configValues
      .filter(v => changeKeys.includes(v.id))
      .map(orig => ({
        id: orig.id,
        oldValue: orig.value,
        value: this.props.fieldValues[orig.id],
        type: orig.type,
      }))

    return (
      <FormDialog
        confirm
        disableGutters
        title={`Apply Configuration Change${changes.length > 1 ? 's' : ''}?`}
        onClose={this.props.onClose}
        onSubmit={() =>
          commit({
            variables: {
              input: changes.map(c => omit(c, ['oldValue', 'type'])),
            },
          })
        }
        errors={nonFieldErrors(error).concat(
          fieldErrors(error).map(e => ({
            message: `${e.field}: ${e.message}`,
          })),
        )}
        form={
          <List data-cy='confirmation-diff'>
            {changes.map(c => (
              <ListItem divider key={c.id} data-cy={'diff-' + c.id}>
                <ListItemText
                  disableTypography
                  secondary={
                    <Diff
                      oldValue={
                        c.type === 'stringList'
                          ? c.oldValue.split(/\n/).join(', ')
                          : c.oldValue
                      }
                      newValue={
                        c.type === 'stringList'
                          ? c.value.split(/\n/).join(', ')
                          : c.value
                      }
                      type={c.type === 'boolean' ? 'words' : 'chars'}
                    />
                  }
                >
                  <Typography>{c.id}</Typography>
                </ListItemText>
              </ListItem>
            ))}
          </List>
        }
      />
    )
  }
}
