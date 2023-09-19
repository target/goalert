import React, { useState, useEffect, SyntheticEvent } from 'react'
import Grid from '@mui/material/Grid'
import { FormContainer, FormField } from '../../forms'
import { FieldError } from '../../util/errutil'
import { CreateGQLAPIKeyInput } from '../../../schema'
import AdminAPIKeyExpirationField from './AdminAPIKeyExpirationField'
import dayjs from 'dayjs'
import { gql, useQuery } from '@apollo/client'
import { GenericError } from '../../error-pages'
import Spinner from '../../loading/components/Spinner'
import { TextField, Autocomplete, MenuItem } from '@mui/material'
import CheckIcon from '@mui/icons-material/Check'

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
}

export default function AdminAPIKeyCreateForm(
  props: AdminAPIKeyCreateFormProps,
): JSX.Element {
  const { ...containerProps } = props
  const [expiresAt, setExpiresAt] = useState<string>(
    dayjs().add(7, 'day').toString(),
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
          <AdminAPIKeyExpirationField
            setValue={setExpiresAt}
            value={expiresAt}
          />
        </Grid>
        <Grid item xs={12}>
          <Autocomplete
            sx={{ m: 1, width: 500 }}
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
          />
        </Grid>
      </Grid>
    </FormContainer>
  )
}
