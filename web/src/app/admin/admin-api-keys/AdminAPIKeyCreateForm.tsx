import React, { useState, useEffect, SyntheticEvent } from 'react'
import Grid from '@mui/material/Grid'
import { FormContainer, FormField } from '../../forms'
import { FieldError } from '../../util/errutil'
import { CreateGQLAPIKeyInput, UserRole } from '../../../schema'
import AdminAPIKeyExpirationField from './AdminAPIKeyExpirationField'
import { gql, useQuery } from '@apollo/client'
import { GenericError } from '../../error-pages'
import Spinner from '../../loading/components/Spinner'
import { TextField, Autocomplete, MenuItem } from '@mui/material'
import CheckIcon from '@mui/icons-material/Check'
import { DateTime } from 'luxon'
import Select, { SelectChangeEvent } from '@mui/material/Select'

const listGQLFieldsQuery = gql`
  query ListGQLFieldsQuery {
    listGQLFields
  }
`
const MaxDetailsLength = 6 * 1024 // 6KiB

interface AdminAPIKeyCreateFormProps {
  value: CreateGQLAPIKeyInput
  errors: FieldError[]
  onChange: (key: CreateGQLAPIKeyInput) => void
  disabled?: boolean
  allowFieldsError: string
}

export default function AdminAPIKeyCreateForm(
  props: AdminAPIKeyCreateFormProps,
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
  const [role, setRole] = useState<UserRole>('user')
  const handleAutocompleteChange = (
    event: SyntheticEvent<Element, Event>,
    value: string[],
  ): void => {
    setAllowedFields(value)
  }

  const handleRoleChange = (event: SelectChangeEvent): void => {
    const val = event.target.value as UserRole
    setRole(val)
  }

  useEffect(() => {
    const valTemp = props.value
    valTemp.expiresAt = new Date(expiresAt).toISOString()
    valTemp.allowedFields = allowedFields
    valTemp.role = role as UserRole

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
            hint='Markdown Supported'
          />
        </Grid>
        <Grid item xs={12}>
          <Select
            labelId='role-select-label'
            id='role-select'
            value={role}
            label='User Role'
            name='userrole'
            onChange={handleRoleChange}
            required
            style={{ width: '100%' }}
          >
            <MenuItem value='user'>User</MenuItem>
            <MenuItem value='admin'>Admin</MenuItem>
          </Select>
        </Grid>
        <Grid item xs={12}>
          <AdminAPIKeyExpirationField
            setValue={setExpiresAt}
            value={expiresAt}
          />
        </Grid>
        <Grid item xs={12}>
          <Autocomplete
            multiple
            options={data.listGQLFields}
            getOptionLabel={(option: string) => option}
            onChange={handleAutocompleteChange}
            disableCloseOnSelect
            renderInput={(params) => (
              <TextField
                {...params}
                variant='outlined'
                label='Allowed Fields'
                placeholder='Allowed Fields'
                helperText={props.allowFieldsError}
                error={props.allowFieldsError !== ''}
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
