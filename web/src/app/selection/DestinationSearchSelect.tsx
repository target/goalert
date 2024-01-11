import React, { useState } from 'react'
import { useQuery, gql } from 'urql'
import {
  DestinationFieldConfig,
  DestinationType,
  FieldValueConnection,
} from '../../schema'
import MaterialSelect from './MaterialSelect'
import { FavoriteIcon } from '../util/SetFavoriteButton'

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
}

const cacheByJSON: Record<string, unknown> = {}

function cachify<T>(val: T): T {
  const json = JSON.stringify(val)
  if (cacheByJSON[json]) return cacheByJSON[json] as T
  cacheByJSON[json] = val

  return val
}

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

  const [{ data: selectedLabelData }] = useQuery<{
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
  const selectedLabel = selectedLabelData?.destinationFieldValueName || ''

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
      isLoading={fetching}
      multiple={false}
      noOptionsText='No options'
      disabled={props.disabled}
      noOptionsError={error}
      onInputChange={(val) => setInputValue(val)}
      value={value as unknown as SelectOption}
      label={props.labelSingular}
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
