import React, { useState } from 'react'
import { useQuery, gql } from 'urql'
import {
  DestinationFieldConfig,
  DestinationType,
  FieldValueConnection,
} from '../../schema'
import MaterialSelect from './MaterialSelect'
import { FavoriteIcon } from '../util/SetFavoriteButton'
import AppLink from '../util/AppLink'

const searchOptionsQuery = gql`
  query DestinationFieldSearch($input: DestinationFieldSearchInput!) {
    destinationFieldSearch(input: $input) {
      nodes {
        value
        label
        isFavorite
      }
    }
  }
`

const selectedLabelQuery = gql`
  query DestinationFieldValueName($input: DestinationFieldValidateInput!) {
    destinationFieldValueName(input: $input)
  }
`

const noSuspense = { suspense: false }

export type DestinationSearchSelectProps = DestinationFieldConfig & {
  value: string
  onChange?: (newValue: string) => void
  destType: DestinationType

  disabled?: boolean
  error?: boolean
}

const cacheByJSON: Record<string, unknown> = {}

function replacer(key: string, value: string): string | undefined {
  if (key === 'icon') return undefined
  return value
}

function cachify<T>(val: T): T {
  const json = JSON.stringify(val, replacer) // needed to avoid circular refs when using rendered icons
  if (cacheByJSON[json]) return cacheByJSON[json] as T
  cacheByJSON[json] = val

  return val
}

/**
 * DestinationSearchSelect is a select field that allows the user to select a
 * destination from a list of options.
 *
 * You should almost never use this component directly. Instead, use
 * DestinationField, which will select the correct component based on the
 * destination type.
 */
export default function DestinationSearchSelect(
  props: DestinationSearchSelectProps,
): JSX.Element {
  const [inputValue, setInputValue] = useState('')

  // check validation of the input phoneNumber through graphql
  const [{ data, fetching, error }] = useQuery<{
    destinationFieldSearch: FieldValueConnection
  }>({
    query: searchOptionsQuery,
    variables: {
      input: {
        destType: props.destType,
        search: inputValue,
        fieldID: props.fieldID,
      },
    },
    requestPolicy: 'cache-first',
    pause: props.disabled,
    context: noSuspense,
  })
  const options = data?.destinationFieldSearch.nodes || []

  const [{ data: selectedLabelData, error: selectedErr }] = useQuery<{
    destinationFieldValueName: string
  }>({
    query: selectedLabelQuery,
    variables: {
      input: {
        destType: props.destType,
        value: props.value,
        fieldID: props.fieldID,
      },
    },
    requestPolicy: 'cache-first',
    pause: !props.value,
    context: noSuspense,
  })

  let selectedLabel = selectedLabelData?.destinationFieldValueName || ''
  if (selectedErr) {
    selectedLabel = `ERROR: ${selectedErr.message}`
  }

  interface SelectOption {
    value: string
    label: string
  }

  function handleChange(val: SelectOption | null): void {
    if (!props.onChange) return

    // should not be possible since multiple is false
    if (Array.isArray(val)) throw new Error('Multiple values not supported')

    props.onChange(val?.value || '')
  }

  const value = props.value
    ? { label: selectedLabel, value: props.value }
    : null

  return (
    <MaterialSelect
      name={props.fieldID}
      isLoading={fetching}
      multiple={false}
      noOptionsText='No options'
      disabled={props.disabled}
      noOptionsError={error}
      error={props.error}
      onInputChange={(val) => setInputValue(val)}
      value={value as unknown as SelectOption}
      label={props.label}
      helperText={
        props.hintURL ? (
          <AppLink newTab to={props.hintURL}>
            {props.hint}
          </AppLink>
        ) : (
          props.hint
        )
      }
      options={options
        .map((opt) => ({
          label: opt.label,
          value: opt.value,
          icon: opt.isFavorite ? <FavoriteIcon /> : undefined,
        }))
        .map(cachify)}
      placeholder='Start typing...'
      onChange={handleChange}
    />
  )
}
