import React from 'react'
import p from 'prop-types'
import Query from '../util/Query'
import { debounce } from 'lodash-es'
import MaterialSelect from './MaterialSelect'
import { mergeFields, fieldAlias, mapInputVars } from '../util/graphql'
import FavoriteIcon from '@material-ui/icons/Star'
import { DEBOUNCE_DELAY } from '../config'
import { POLL_INTERVAL } from '../util/poll_intervals'

// valueCheck ensures the type is `arrayOf(p.string)` if `multiple` is set
// and `p.string` otherwise.
function valueCheck(props, ...args) {
  if (props.multiple) return p.arrayOf(p.string).isRequired(props, ...args)
  return p.string(props, ...args)
}

/*
 * This component is to be used when we want to create a select based on a query from the database.
 * (i.e. users, services, schedules, etc)
 */
export default class QuerySelect extends React.PureComponent {
  static propTypes = {
    // The provided query must map to a Search/Connection type
    // field. (like `escalationPolicies` or `users`)
    query: p.object.isRequired,

    // valueQuery should return the same data as `query` but for
    // a single element given an `id` input variable.
    valueQuery: p.object,

    // mapDataNode should take a node from the query response and return
    // a structure in the format of: `{label, value, isFavorite}`
    mapDataNode: p.func,

    // variables can be used to add extra parameters to the query
    // such as: `{ input: { favoritesFirst: true } }`
    //
    // Provided variables will be merged with normal pagination/search.
    variables: p.object,

    // If defaultQueryVariables is set, then when there is no search
    // (user has focused the select, but not typed anything) it will be
    // used to fetch a list of default results.
    //
    // This can be used to display favorites by default by setting
    // `defaultQueryVariables={ input: { favoritesOnly: true } }`
    defaultQueryVariables: p.object,

    // If specified, a "Create" option will be provided for the users
    // provided text. It will be called instead of `onChange` if the user
    // selects the generated option.
    //
    // Example: if the user types `foobar` and there is not a `foobar` option,
    // then `Create "foobar"` will be displayed in the dropdown.
    onCreate: p.func,

    error: p.bool,
    onChange: p.func.isRequired,
    value: valueCheck,

    multiple: p.bool,
    name: p.string,
    placeholder: p.string,
  }

  static defaultProps = {
    error: false,
    value: null,
    multiple: false,
    mapDataNode: node => ({
      label: node.name,
      value: node.id,
      isFavorite: Boolean(node.isFavorite),
    }),
    variables: {},
  }

  state = {
    outdated: false,
    search: '',
    skip: true,
  }

  componentDidMount() {
    this._refresh = setInterval(() => this.forceUpdate(), POLL_INTERVAL)
  }
  componentWillUnmount() {
    clearInterval(this._refresh)
    this.onInputChange.cancel()
  }

  onInputChange = debounce(search => {
    this.setState({ search, skip: !search, outdated: false })
  }, DEBOUNCE_DELAY)

  isEmpty() {
    return this.props.multiple
      ? this.props.value.length === 0
      : !this.props.value
  }

  render() {
    return this.renderValueQuery()
  }

  renderValueQuery() {
    if (!this.props.valueQuery) {
      // if no query provided, display the value itself
      const mapVal = v => ({ value: v, label: v })
      let value = []
      if (this.props.value && this.props.multiple) {
        value = this.props.value.map(mapVal)
      } else if (this.props.value && !this.props.multiple) {
        value = [mapVal(this.props.value)]
      }

      return this.renderOptionsQuery(value)
    }

    let query, variables, getOptions
    if (this.props.multiple) {
      query = fieldAlias(this.props.valueQuery, 'data')
      variables = { id: this.props.value[0] }
      this.props.value.slice(1).forEach((val, idx) => {
        const varName = 'id' + idx

        query = mergeFields(
          query,
          mapInputVars(fieldAlias(this.props.valueQuery, 'data' + idx), {
            id: varName,
          }),
        )

        variables[varName] = val
      })
      getOptions = data =>
        data && data.data ? Object.values(data).map(this.props.mapDataNode) : []
    } else {
      query = fieldAlias(this.props.valueQuery, 'data')
      variables = { id: this.props.value }
      getOptions = data =>
        data && data.data ? [this.props.mapDataNode(data.data)] : []
    }

    return (
      <Query
        query={query}
        variables={variables}
        skip={this.isEmpty()}
        noError
        noSpin
        noPoll
        // rely on cache, search results will handle updating it if need-be
        fetchPolicy='cache-first'
        render={({ data }) => this.renderOptionsQuery(getOptions(data))}
      />
    )
  }

  renderOptionsQuery(valueOptions) {
    const getOptions = data =>
      data && data.data && data.data.nodes
        ? data.data.nodes.map(this.props.mapDataNode)
        : []

    const skip = Boolean(
      (this.state.skip && this.state.search) ||
        (!this.state.search && !this.props.defaultQueryVariables),
    )
    const omitOpts = {}
    if (this.props.multiple) {
      omitOpts.omit = this.props.value
    }

    let variables
    if (!this.state.search && this.props.defaultQueryVariables) {
      variables = this.props.defaultQueryVariables
    } else {
      variables = {
        ...this.props.variables,
        input: {
          first: 5,
          search: this.state.search,
          ...omitOpts,
          ...this.props.variables.input,
        },
      }
    }

    return (
      <Query
        query={fieldAlias(this.props.query, 'data')}
        skip={skip}
        noError
        noSpin
        fetchPolicy='network-only'
        variables={variables}
        render={({ data, loading, error }) =>
          this.renderSelect(valueOptions, {
            options: getOptions(data),
            loading,
            error,
          })
        }
      />
    )
  }

  renderSelect(valueOptions, { options, loading, error }) {
    // keep unused variables so they are not used in rest spread
    const {
      onCreate,
      query,
      valueQuery,
      mapDataNode,
      onChange,
      value,
      search,
      variables,
      defaultQueryVariables,
      placeholder,
      ...rest
    } = this.props

    let noOptionsMessage = 'No options'
    if (this.state.skip) noOptionsMessage = 'Start typing...'
    else if (error) noOptionsMessage = 'Error: ' + error

    const selectOptions =
      this.state.search || defaultQueryVariables
        ? options.concat(valueOptions)
        : valueOptions
    let selectValue, changeCallback

    const mapVal = val =>
      selectOptions.find(opt => opt.value === val) || {
        value: val,
        label: 'Loading...',
      }

    if (this.props.multiple) {
      changeCallback = val => {
        const created = val.find(v => v.isCreate)
        if (created) {
          onCreate(created.value)
          return
        }
        onChange(val ? val.map(v => v.value) : [])
      }
      selectValue = value.map(mapVal)
    } else {
      changeCallback = val => {
        if (val && val.isCreate) {
          onCreate(val.value)
          return
        }
        onChange((val && val.value) || '')
      }
      if (value) selectValue = mapVal(value)
    }

    if (
      onCreate &&
      this.state.search &&
      !selectOptions.find(o => o.value === this.state.search)
    ) {
      selectOptions.push({
        isCreate: true,
        value: this.state.search,
        label: `Create "${this.state.search}"`,
      })
    }

    return (
      <MaterialSelect
        isLoading={this.state.outdated || (loading && this.state.search)}
        noOptionsMessage={() => noOptionsMessage}
        onInputChange={val => {
          this.setState({ outdated: true })
          this.onInputChange(val)
        }}
        value={selectValue}
        options={
          (this.state.outdated && []) ||
          selectOptions.map(opt => ({
            ...opt,
            icon: opt.isFavorite ? <FavoriteIcon /> : null,
          }))
        }
        placeholder={
          placeholder ||
          (defaultQueryVariables && !search ? 'Start typing...' : null)
        }
        onChange={val => {
          this.setState({ search: '', outdated: false })

          // ignore any pending search updates
          this.onInputChange.cancel()
          changeCallback(val)
        }}
        {...rest}
      />
    )
  }
}
