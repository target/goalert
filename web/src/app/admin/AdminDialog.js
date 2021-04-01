import React from 'react'
import p from 'prop-types'
import List from '@material-ui/core/List'
import ListItem from '@material-ui/core/ListItem'
import ListItemText from '@material-ui/core/ListItemText'
import Typography from '@material-ui/core/Typography'
import { omit } from 'lodash-es'
import FormDialog from '../dialogs/FormDialog'
import { nonFieldErrors, fieldErrors } from '../util/errutil'
import Diff from '../util/Diff'
import { useMutation } from '@apollo/client'

function AdminDialog(props) {
  const [commit, { error }] = useMutation(props.mutation, {
    onCompleted: props.onComplete,
  })
  const changeKeys = Object.keys(props.fieldValues)
  const changes = props.values
    .filter((v) => changeKeys.includes(v.id))
    .map((orig) => ({
      id: orig.id,
      oldValue: orig.value,
      value: props.fieldValues[orig.id],
      type: orig.type || typeof orig.value,
    }))

  return (
    <FormDialog
      disableGutters
      title={`Apply Configuration Change${changes.length > 1 ? 's' : ''}?`}
      onClose={props.onClose}
      onSubmit={() =>
        commit({
          variables: {
            input: changes.map((c) => {
              c.value = c.value === '' && c.type === 'number' ? '0' : c.value
              return omit(c, ['oldValue', 'type'])
            }),
          },
        })
      }
      primaryActionLabel='Confirm'
      errors={nonFieldErrors(error).concat(
        fieldErrors(error).map((e) => ({
          message: `${e.field}: ${e.message}`,
        })),
      )}
      form={
        <List data-cy='confirmation-diff'>
          {changes.map((c) => (
            <ListItem divider key={c.id} data-cy={'diff-' + c.id}>
              <ListItemText
                disableTypography
                secondary={
                  <Diff
                    oldValue={
                      c.type === 'stringList'
                        ? c.oldValue.split(/\n/).join(', ')
                        : c.oldValue.toString()
                    }
                    newValue={
                      c.type === 'stringList'
                        ? c.value.split(/\n/).join(', ')
                        : c.value.toString()
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

AdminDialog.propTypes = {
  mutation: p.object.isRequired,
  values: p.array.isRequired,
  fieldValues: p.object.isRequired,
  onClose: p.func.isRequired,
  onComplete: p.func.isRequired,
}

export default AdminDialog
