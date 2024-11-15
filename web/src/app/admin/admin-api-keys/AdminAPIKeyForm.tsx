import React, { lazy } from 'react'
import Grid from '@mui/material/Grid'
import { FormContainer, FormField } from '../../forms'
import { FieldError } from '../../util/errutil'
import { CreateGQLAPIKeyInput } from '../../../schema'
import AdminAPIKeyExpirationField from './AdminAPIKeyExpirationField'
import { TextField, MenuItem } from '@mui/material'

const GraphQLEditor = lazy(() => import('../../editor/GraphQLEditor'))

type AdminAPIKeyFormProps = {
  errors: FieldError[]

  // even while editing, we need all the fields
  value: CreateGQLAPIKeyInput
  onChange: (key: CreateGQLAPIKeyInput) => void

  create?: boolean
}

export default function AdminAPIKeyForm(
  props: AdminAPIKeyFormProps,
): JSX.Element {
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
          <GraphQLEditor
            value={props.value.query}
            onChange={(query) => props.onChange({ ...props.value, query })}
            minHeight='20em'
          />
        </Grid>
      </Grid>
    </FormContainer>
  )
}
