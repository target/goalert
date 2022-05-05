import React from 'react'
import List from '@mui/material/List'
import ListItem from '@mui/material/ListItem'
import ListItemText from '@mui/material/ListItemText'
import Typography from '@mui/material/Typography'
import { omit } from 'lodash'
import FormDialog from '../dialogs/FormDialog'
import { nonFieldErrors, fieldErrors } from '../util/errutil'
import Diff from '../util/Diff'
import { DocumentNode, useMutation } from '@apollo/client'
import { ConfigValue } from '../../schema'

interface FieldValues {
  [id: string]: string
}

interface AdminDialogProps {
  mutation: DocumentNode
  values: ConfigValue[]
  fieldValues: FieldValues
  onClose: () => void
  onComplete: () => void
}

function AdminDialog(props: AdminDialogProps): JSX.Element {
  const [commit, { error }] = useMutation(props.mutation, {
    onCompleted: props.onComplete,
  })
  const changeKeys = Object.keys(props.fieldValues)
  const changes = props.values
    .filter((v: { id: string }) => changeKeys.includes(v.id))
    .map((orig) => ({
      id: orig.id,
      oldValue: orig.value,
      value: props.fieldValues[orig.id],
      type: orig.type || typeof orig.value,
    }))

  const nonFieldErrs = nonFieldErrors(error).map((e) => ({
    message: e.message,
  }))
  const fieldErrs = fieldErrors(error).map((e) => ({
    message: `${e.field}: ${e.message}`,
  }))
  const errs = nonFieldErrs.concat(fieldErrs)

  return (
    <FormDialog
      title={`Apply Configuration Change${changes.length > 1 ? 's' : ''}?`}
      onClose={props.onClose}
      onSubmit={() =>
        commit({
          variables: {
            input: changes.map((c: { value: string; type: string }) => {
              c.value = c.value === '' && c.type === 'number' ? '0' : c.value
              return omit(c, ['oldValue', 'type'])
            }),
          },
        })
      }
      primaryActionLabel='Confirm'
      errors={errs}
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

export default AdminDialog
