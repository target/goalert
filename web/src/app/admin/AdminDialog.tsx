import React from 'react'
import List from '@mui/material/List'
import ListItem from '@mui/material/ListItem'
import ListItemText from '@mui/material/ListItemText'
import Typography from '@mui/material/Typography'
import { omit } from 'lodash'
import FormDialog from '../dialogs/FormDialog'
import { nonFieldErrors, fieldErrors } from '../util/errutil'
import Diff from '../util/Diff'
import { ConfigValue } from '../../schema'
import { gql, useMutation, useQuery } from 'urql'
import { DocumentNode } from 'graphql'

const query = gql`
  query {
    values: config(all: true) {
      id
      value
    }
  }
`

const mutation = gql`
  mutation ($input: [ConfigValueInput!]) {
    setConfig(input: $input)
  }
`

interface AdminDialogProps {
  query?: DocumentNode
  mutation?: DocumentNode
  value: { [id: string]: string }
  onClose: () => void
  onComplete?: () => void
}

function AdminDialog(props: AdminDialogProps): React.JSX.Element {
  const [{ data, fetching, error: readError }] = useQuery({
    query: props.query || query,
  })
  const [{ error }, commit] = useMutation(props.mutation || mutation)

  const currentConfig: ConfigValue[] = data?.values || []

  const changeKeys = Object.keys(props.value)
  const changes = currentConfig
    .filter(
      (v: ConfigValue) =>
        changeKeys.includes(v.id) && v.value !== props.value[v.id],
    )
    .map((orig) => ({
      id: orig.id,
      oldValue: orig.value,
      value: props.value[orig.id],
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
        commit(
          {
            input: changes.map((c: { value: string; type: string }) => {
              c.value = c.value === '' && c.type === 'number' ? '0' : c.value
              return omit(c, ['oldValue', 'type'])
            }),
          },
          { additionalTypenames: ['ConfigValue', 'SystemLimit'] },
        ).then((res) => {
          if (res.error) return

          if (props.onComplete) props.onComplete()
          props.onClose()
        })
      }
      primaryActionLabel='Confirm'
      loading={fetching}
      errors={errs.concat(readError ? [readError] : [])}
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
          {changes.length === 0 && (
            <Typography>No changes to apply, already configured.</Typography>
          )}
        </List>
      }
    />
  )
}

export default AdminDialog
