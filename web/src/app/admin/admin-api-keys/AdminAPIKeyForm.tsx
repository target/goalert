import React, { useState, useEffect, SyntheticEvent } from 'react'
import Grid from '@mui/material/Grid'
import { FormContainer, FormField } from '../../forms'
import { FieldError } from '../../util/errutil'
import { GQLAPIKey } from '../../../schema'
import AdminAPIKeyExpirationField from './AdminAPIKeyExpirationField'
import { gql, useQuery } from '@apollo/client'
import { GenericError } from '../../error-pages'
import Spinner from '../../loading/components/Spinner'
import { TextField, Autocomplete, MenuItem } from '@mui/material'
import CheckIcon from '@mui/icons-material/Check'
import { DateTime } from 'luxon'

const listGQLFieldsQuery = gql`
  query ListGQLFieldsQuery {
    listGQLFields
  }
`
const MaxDetailsLength = 6 * 1024 // 6KiB

interface AdminAPIKeyFormProps {
  value: GQLAPIKey
  errors: FieldError[]
  onChange: (key: GQLAPIKey) => void
  disabled?: boolean
  allowFieldsError: boolean
  setAllowFieldsError: (param: boolean) => void
  create: boolean
}

export default function AdminAPIKeyForm(
  props: AdminAPIKeyFormProps,
): JSX.Element {
  const { ...containerProps } = props
  const [expiresAt, setExpiresAt] = useState<string>(
    DateTime.now().plus({ days: 7 }).toLocaleString({
      weekday: 'short',
      month: 'short',
      day: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    }),
  )
  const [allowedFields, setAllowedFields] = useState<string[]>([])
  const handleAutocompleteChange = (
    event: SyntheticEvent<Element, Event>,
    value: string[],
  ): void => {
    setAllowedFields(value)
  }

  useEffect(() => {
    const valTemp = props.value
    valTemp.expiresAt = new Date(expiresAt).toISOString()
    valTemp.allowedFields = allowedFields
    props.setAllowFieldsError(valTemp.allowedFields.length <= 0)
    props.onChange(valTemp)
  })
  const { data, loading, error } = useQuery(listGQLFieldsQuery)

  if (error) {
    return <GenericError error={error.message} />
  }

  if (loading && !data) {
    return <Spinner />
  }

  return (
    <FormContainer {...containerProps}>
      <Grid container spacing={2}>
        <Grid item xs={12}>
          <FormField
            fullWidth
            label='Name'
            name='name'
            required
            component={TextField}
            value={props.value.name}
          />
        </Grid>
        <Grid item xs={12}>
          <FormField
            fullWidth
            label='Description'
            name='description'
            multiline
            rows={4}
            required
            component={TextField}
            charCount={MaxDetailsLength}
            value={props.value.description}
            hint='Markdown Supported'
          />
        </Grid>
        <Grid item xs={12}>
          <FormField
            fullWidth
            component={TextField}
            select
            required
            label='Role'
            name='role'
            disabled={!props.create}
            value={props.value.role}
          >
            <MenuItem value='user' key='user'>
              User
            </MenuItem>
            <MenuItem value='admin' key='user'>
              Admin
            </MenuItem>
          </FormField>
        </Grid>
        <Grid item xs={12}>
          {props.create ? (
            <AdminAPIKeyExpirationField
              setValue={setExpiresAt}
              value={expiresAt}
              create
            />
          ) : (
            <FormField
              fullWidth
              label='Expires At'
              name='expiresAt'
              component={TextField}
              value={props.value.expiresAt}
              disabled
            />
          )}
        </Grid>
        <Grid item xs={12}>
          <Autocomplete
            multiple
            options={data.listGQLFields}
            getOptionLabel={(option: string) => option}
            onChange={handleAutocompleteChange}
            disabled={!props.create}
            disableCloseOnSelect
            defaultValue={props.value.allowedFields}
            renderInput={(params) => (
              <FormField
                {...params}
                fullWidth
                label='allowedFields'
                name='Allowed Fields'
                required={props.allowFieldsError && props.create}
                component={TextField}
                disabled={!props.create}
              />
            )}
            renderOption={(props, option, { selected }) => (
              <MenuItem
                {...props}
                key={option}
                value={option}
                sx={{ justifyContent: 'space-between' }}
              >
                {option}
                {selected ? <CheckIcon color='info' /> : null}
              </MenuItem>
            )}
            style={{ width: '100%' }}
          />
        </Grid>
      </Grid>
    </FormContainer>
  )
}
