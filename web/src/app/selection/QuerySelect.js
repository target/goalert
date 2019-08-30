import React, { useState, useEffect } from 'react'
import p from 'prop-types'
import { memoize, omit } from 'lodash-es'
import MaterialSelect from './MaterialSelect'
import { mergeFields, fieldAlias, mapInputVars } from '../util/graphql'
import FavoriteIcon from '@material-ui/icons/Star'
import { DEBOUNCE_DELAY } from '../config'
import { useQuery } from 'react-apollo'

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

const asArray = value => {
  if (!value) return []

  return Array.isArray(value) ? value : [value]
}
const mapValueQuery = (query, index) =>
  mapInputVars(fieldAlias(query, 'data' + index), { id: 'id' + index })
// useValues will return a set of {id, label}
// for the provided value.
//
// If value is not an array, a single object (instead of array)
// is returned.
function makeUseValues(query, mapNode) {
  if (!query) {
    // no value query, so always use the map function
    return function useValuesNoQuery(_value) {
      const value = asArray(_value).map(v => ({ value: v, label: v }))
      return [Array.isArray(_value) ? value : value[0] || null, null]
    }
  }

  const getQueryBySize = memoize(size => {
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

    const { data, error } = useQuery(getQueryBySize(value.length), {
      skip: !value.length,
      variables,
      returnPartialData: true,
      fetchPolicy: 'cache-first',
      pollInterval: 0,
    })

    if (!value.length) {
      return [null, error]
    }

    const result = value.map((v, i) => {
      const name = 'data' + i
      if (!data || !data[name]) return { value: v, label: 'Loading...' }

      return mapNode(data[name])
    })

    if (Array.isArray(_value)) {
      return [result, error]
    }
    return [result[0], error]
  }
}

// makeUseOptions will provide the available options for the given query.
function makeUseOptions(query, mapNode, vars, defaultVars) {
  const q = fieldAlias(query, 'data')
  return function useOptions(value, search) {
    const params = { first: 5, omit: asArray(value) }
    const input = search
      ? { ...vars, ...params, search }
      : { ...defaultVars, ...params }

    const { data, loading, error } = useQuery(q, {
      skip: !search && !defaultVars,
      variables: { input },
      fetchPolicy: 'network-only',
      pollInterval: 0,
    })

    let result = []

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
  onChange: p.func,
  value: valueCheck,

  multiple: p.bool,
  name: p.string,
  placeholder: p.string,
}

export function makeQuerySelect(displayName, options) {
  const {
    mapDataNode = defaultMapNode,
    variables = {},
    query,
    valueQuery,
    defaultQueryVariables,
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

      onCreate = () => {},
      onChange = () => {},
      ...otherProps
    } = props

    const [search, setSearch] = useState('')
    const [searchInput, setSearchInput] = useState('')
    const [renderCheck, setRenderCheck] = useState(0)
    const [optionCache] = useState({})
    const [selectValue] = useValues(value)
    const [
      selectOptions,
      { loading: optionsLoading, error: optionsError },
    ] = useOptions(value, search)

    useEffect(() => {
      const t = setTimeout(() => setRenderCheck(renderCheck + 1), 1000)
      return () => clearTimeout(t)
    })
    useEffect(() => {
      const t = setTimeout(() => setSearch(searchInput), DEBOUNCE_DELAY)

      return () => clearTimeout(t)
    }, [searchInput])

    const cachify = option => {
      const key = JSON.stringify(omit(option, 'icon'))
      if (!optionCache[key]) optionCache[key] = option
      return optionCache[key]
    }

    const handleChange = newVal => {
      setSearch('')
      setSearchInput('')
      const created = asArray(newVal).find(v => v.isCreate)
      if (created) onCreate(created.value)
      else if (multiple) onChange(asArray(newVal).map(v => v.value))
      else onChange(newVal.value)
    }

    let noOptionsMessage = 'No options'
    if (!searchInput && !selectOptions.length)
      noOptionsMessage = 'Start typing...'
    else if (optionsError) noOptionsMessage = 'Error: ' + optionsError

    return (
      <MaterialSelect
        isLoading={search !== searchInput || optionsLoading}
        noOptionsMessage={() => noOptionsMessage}
        onInputChange={val => setSearchInput(val)}
        value={multiple ? asArray(selectValue) : selectValue}
        multiple={multiple}
        options={selectOptions
          .map(opt => ({
            ...opt,
            icon: opt.isFavorite ? <FavoriteIcon /> : null,
          }))
          .map(cachify)}
        placeholder={
          placeholder ||
          (defaultQueryVariables && !searchInput ? 'Start typing...' : null)
        }
        onChange={val => handleChange(val)}
        {...otherProps}
      />
    )
  }

  QuerySelect.displayName = displayName
  QuerySelect.propTypes = querySelectPropTypes

  return QuerySelect
}
