import React, { useEffect, useState } from 'react'
import Grid from '@mui/material/Grid'
import { FormContainer, FormField } from '../../forms'
import { FieldError } from '../../util/errutil'
import { CreateGQLAPIKeyInput } from '../../../schema'
import AdminAPIKeyExpirationField from './AdminAPIKeyExpirationField'
import { gql, useQuery } from 'urql'
import { GenericError } from '../../error-pages'
import Spinner from '../../loading/components/Spinner'
import { TextField, MenuItem } from '@mui/material'
import MaterialSelect from '../../selection/MaterialSelect'
import ClickableText from '../../util/ClickableText'
import CompareArrows from '@mui/icons-material/CompareArrows'

const query = gql`
  query ListGQLFieldsQuery {
    listGQLFields
  }
`

const queryFields = gql`
  query ListExampleFieldsQuery($query: String!) {
    listGQLFields(query: $query)
  }
`

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
  const [showQuery, setShowQuery] = useState(false)
  const [exampleQuery, setExampleQuery] = useState('')

  const [{ data, fetching, error }] = useQuery({
    query,
  })

  const [example] = useQuery({
    query: queryFields,
    pause: !showQuery || !exampleQuery,
    variables: {
      query: exampleQuery,
    },
  })
  const exampleFields = example?.data?.listGQLFields || []
  const exampleLoaded = !example?.fetching && !example?.error

  useEffect(() => {
    if (!showQuery) return
    if (!exampleQuery) return
    if (!exampleLoaded) return

    props.onChange({ ...props.value, allowedFields: exampleFields })
  }, [exampleFields, showQuery, exampleQuery, exampleLoaded])

  if (error) {
    return <GenericError error={error.message} />
  }

  if (fetching && !data) {
    return <Spinner />
  }

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
          {showQuery && (
            <TextField
              fullWidth
              multiline
              label='Example Query'
              placeholder='Enter GraphQL query here...'
              value={exampleQuery}
              onChange={(e) => setExampleQuery(e.target.value)}
              error={!!example?.error}
              helperText={
                <React.Fragment>
                  <div>{example?.error?.message}</div>
                  <ClickableText
                    onClick={() => setShowQuery(false)}
                    endIcon={<CompareArrows />}
                  >
                    Select fields manually
                  </ClickableText>
                </React.Fragment>
              }
            />
          )}
          {!showQuery && (
            <FormField
              fullWidth
              component={MaterialSelect}
              name='allowedFields'
              disabled={!props.create}
              clientSideFilter
              disableCloseOnSelect
              optionsLimit={10}
              options={data.listGQLFields.map((field: string) => ({
                label: field,
                value: field,
              }))}
              mapOnChangeValue={(selected: { value: string }[]) =>
                selected.map((v) => v.value)
              }
              mapValue={(value: string[]) =>
                value.map((v) => ({ label: v, value: v }))
              }
              multiple
              required
              hint={
                <ClickableText
                  onClick={() => setShowQuery(true)}
                  endIcon={<CompareArrows />}
                >
                  Enter example query instead
                </ClickableText>
              }
            />
          )}
        </Grid>
      </Grid>
    </FormContainer>
  )
}
