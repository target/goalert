import React from 'react'
import Grid from '@mui/material/Grid'
import { FormContainer, FormField, HelperText } from '../../forms'
import { FieldError } from '../../util/errutil'
import { CreateGQLAPIKeyInput } from '../../../schema'
import AdminAPIKeyExpirationField from './AdminAPIKeyExpirationField'
import { TextField, MenuItem, FormControl } from '@mui/material'
import GraphQLEditor from '../../editor/GraphQLEditor'

type AdminAPIKeyFormProps = {
  errors: FieldError[]

  // even while editing, we need all the fields
  value: CreateGQLAPIKeyInput
  onChange: (key: CreateGQLAPIKeyInput) => void

  create?: boolean
}

export default function AdminAPIKeyForm(
  props: AdminAPIKeyFormProps,
): React.JSX.Element {
  const queryError = props.errors.find((e) => e.field === 'query')?.message
  return (
    <FormContainer optionalLabels {...props}>
      <Grid container spacing={2}>
        <Grid item xs={12}>
          <FormField fullWidth name='name' required component={TextField} />
        </Grid>
        <Grid item xs={12}>
          <FormField
            fullWidth
            name='description'
            multiline
            rows={4}
            required
            component={TextField}
            charCount={255}
            hint='Markdown Supported'
          />
        </Grid>
        <Grid item xs={12}>
          <FormField
            fullWidth
            component={TextField}
            select
            required
            name='role'
            disabled={!props.create}
          >
            <MenuItem value='user' key='user'>
              User
            </MenuItem>
            <MenuItem value='admin' key='admin'>
              Admin
            </MenuItem>
          </FormField>
        </Grid>
        <Grid item xs={12}>
          <FormField
            fullWidth
            component={AdminAPIKeyExpirationField}
            select
            required
            name='expiresAt'
            disabled={!props.create}
          />
        </Grid>
        <Grid item xs={12}>
          <FormControl error={!!queryError} fullWidth>
            <GraphQLEditor
              value={props.value.query}
              readOnly={!props.create}
              onChange={(query) => props.onChange({ ...props.value, query })}
              minHeight='20em'
              maxHeight='20em'
            />
            <HelperText
              hint={props.create ? '' : '(read-only)'}
              error={props.errors.find((e) => e.field === 'query')?.message}
            />
          </FormControl>
        </Grid>
      </Grid>
    </FormContainer>
  )
}
