import React, { useState, useEffect } from 'react'
import p from 'prop-types'
import _, { memoize, omit } from 'lodash'
import MaterialSelect from './MaterialSelect'
import { mergeFields, fieldAlias, mapInputVars } from '../util/graphql'
import { DEBOUNCE_DELAY } from '../config'
import { useQuery } from 'urql'
import { FavoriteIcon } from '../util/SetFavoriteButton'
import { print } from 'graphql/language/printer'

// valueCheck ensures the type is `arrayOf(p.string)` if `multiple` is set
// and `p.string` otherwise.
function valueCheck(props, ...args) {
  if (props.multiple) return p.arrayOf(p.string).isRequired(props, ...args)
  return p.string(props, ...args)
}

const defaultMapNode = ({ name: label, id: value, isFavorite }) => ({
  label,
  value,
  isFavorite,
})

const asArray = (value) => {
  if (!value) return []

  return Array.isArray(value) ? value : [value]
}

const mapValueQuery = (query, index) =>
  mapInputVars(fieldAlias(query, 'data' + index), { id: 'id' + index })

// makeUseValues will return a hook that will fetch select values for the
// given ids.
function makeUseValues(query, mapNode) {
  if (!query) {
    // no value query, so always use the map function
    return function useValuesNoQuery(_value) {
      const value = asArray(_value).map((v) => ({ value: v, label: v }))
      return [Array.isArray(_value) ? value : value[0] || null, null]
    }
  }

  const getQueryBySize = memoize((size) => {
    let q = mapValueQuery(query, 0)

    for (let i = 1; i < size; i++) {
      q = mergeFields(q, mapValueQuery(query, i))
    }

    return q
  })
  return function useValuesQuery(_value) {
    const value = asArray(_value)
    const variables = {}
    value.forEach((v, i) => {
      variables['id' + i] = v
    })
    const [{ data, error }] = useQuery({
      query: print(getQueryBySize(value.length)),
      pause: !value.length,
      variables,
      requestPolicy: 'cache-first',
    })

    if (!value.length) {
      return [null, error]
    }

    const result = value.map((v, i) => {
      const name = 'data' + i
      if (!data || _.isEmpty(data[name]))
        return { value: v, label: 'Loading...' }

      return mapNode(data[name])
    })

    if (Array.isArray(_value)) {
      return [result, error]
    }
    return [result[0], error]
  }
}

// makeUseOptions will return a hook that will fetch available options
// for a given search query.
function makeUseOptions(query, mapNode, vars, defaultVars) {
  const q = fieldAlias(query, 'data')
  return function useOptions(value, search, extraVars) {
    const params = { first: 5, omit: Array.isArray(value) ? value : [] } // only omit in multi-select mode
    const input = search
      ? { ...vars, ...extraVars, ...params, search }
      : { ...defaultVars, ...extraVars, ...params }

    const [{ data, fetching: loading, error }] = useQuery({
      query: print(q),
      pause: !search && !defaultVars,
      variables: { input },
      requestPolicy: 'network-only',

      pollInterval: 0,
      errorPolicy: 'all', // needs to be set explicitly for some reason
    })

    let result = []

    if (error) {
      return [[], { error }]
    }
    if (!loading && data && data.data) {
      result = data.data.nodes.map(mapNode)
    }

    return [result, { loading, error }]
  }
}

export const querySelectPropTypes = {
  // If specified, a "Create" option will be provided for the users
  // provided text. It will be called instead of `onChange` if the user
  // selects the generated option.
  //
  // Example: if the user types `foobar` and there is not a `foobar` option,
  // then `Create "foobar"` will be displayed in the dropdown.
  onCreate: p.func,

  error: p.bool,
  formatInputOnChange: p.func,
  onChange: p.func,
  value: valueCheck,
  fullWidth: p.bool,
  label: p.string,
  multiple: p.bool,
  name: p.string,
  placeholder: p.string,
  disabled: p.bool,
  labelKey: p.string,
}

// makeQuerySelect will return a new React component that can be used
// as a select input with connected search.
export function makeQuerySelect(displayName, options) {
  const {
    // mapDataNode can be passed to change how server-returned values are mapped to select options
    mapDataNode = defaultMapNode,

    // variables will be mixed into the query parameters
    variables = {},

    // extraVariablesFunc is an optional function that is passed props
    // and should return a new set of props and extra variables.
    extraVariablesFunc = (props) => [props, {}],

    // query is used to fetch available options. It should define
    // an `input` variable that takes `omit`, `first`, and `search` properties.
    query,

    // valueQuery is optional and used to fetch labels for selected ids.
    valueQuery,

    // defaultQueryVariables, if specified, will be used when the component is active
    // but has no search parameter. It is useful for showing favorites before the user
    // enters a search term.
    defaultQueryVariables = {},
  } = options

  const useValues = makeUseValues(valueQuery, mapDataNode)
  const useOptions = makeUseOptions(
    query,
    mapDataNode,
    variables,
    defaultQueryVariables,
  )

  function QuerySelect(props) {
    const {
      value = props.multiple ? [] : null,
      multiple = false,

      placeholder,

      onCreate: _onCreate,
      onChange = () => {},
      formatInputOnChange = (val) => val,

      ..._otherProps
    } = props

    const [otherProps, extraVars] = extraVariablesFunc(_otherProps)
    const onCreate = _onCreate || (() => {})

    const [search, setSearch] = useState('')
    const [searchInput, setSearchInput] = useState('')
    const [optionCache] = useState({})
    const [selectValue] = useValues(value)
    const [selectOptions, { loading: optionsLoading, error: optionsError }] =
      useOptions(value, search, extraVars)

    if (
      _onCreate &&
      searchInput &&
      !selectOptions.some((o) => o.value === searchInput)
    ) {
      selectOptions.push({
        isCreate: true,
        value: searchInput,
        label: `Create "${searchInput}"`,
      })
    }

    useEffect(() => {
      const t = setTimeout(() => setSearch(searchInput), DEBOUNCE_DELAY)

      return () => clearTimeout(t)
    }, [searchInput])

    const cachify = (option) => {
      const key = JSON.stringify(omit(option, 'icon'))
      if (!optionCache[key]) optionCache[key] = option
      return optionCache[key]
    }

    const handleChange = (newVal) => {
      setSearch('')
      setSearchInput('')
      const created = asArray(newVal).find((v) => v.isCreate)
      if (created) onCreate(created.value)
      else if (multiple) onChange(asArray(newVal).map((v) => v.value))
      else onChange((newVal && newVal.value) || null)
    }

    let noOptionsText = 'No options'
    if (searchInput && selectOptions.length) noOptionsText = 'Start typing...'

    return (
      <MaterialSelect
        isLoading={search !== searchInput || optionsLoading}
        noOptionsText={noOptionsText}
        noOptionsError={optionsError}
        onInputChange={(val) => setSearchInput(val)}
        value={multiple ? asArray(selectValue) : selectValue}
        multiple={multiple}
        options={selectOptions
          .map((opt) => ({
            ...opt,
            icon: opt.isFavorite ? <FavoriteIcon /> : null,
          }))
          .map(cachify)}
        placeholder={
          placeholder ||
          (defaultQueryVariables && !searchInput ? 'Start typing...' : null)
        }
        formatInputOnChange={(val) => formatInputOnChange(val)}
        onChange={(val) => handleChange(val)}
        {...otherProps}
      />
    )
  }

  QuerySelect.displayName = displayName
  QuerySelect.propTypes = querySelectPropTypes

  return QuerySelect
}
