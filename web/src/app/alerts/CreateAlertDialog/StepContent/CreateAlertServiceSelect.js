import React, { useRef, useState, useEffect } from 'react'
import {
  TextField,
  InputAdornment,
  List,
  ListItem,
  ListItemText,
  ListItemIcon,
  Paper,
  Typography,
  Chip,
  FormHelperText,
  FormLabel,
  FormControl,
  Box,
} from '@material-ui/core'
import { makeStyles } from '@material-ui/core/styles'
import ServiceLabelFilterContainer from '../../../services/ServiceLabelFilterContainer'
import { Search as SearchIcon } from '@material-ui/icons'
import FavoriteIcon from '@material-ui/icons/Star'
import { ServiceChip } from '../../../util/Chips'
import AddIcon from '@material-ui/icons/Add'
import _ from 'lodash-es'
import getServiceLabel from '../../../util/getServiceLabel'
import { CREATE_ALERT_LIMIT, DEBOUNCE_DELAY } from '../../../config'
import { useQuery } from 'react-apollo'
import gql from 'graphql-tag'
import { allErrors } from '../../../util/errutil'

const query = gql`
  query($input: ServiceSearchOptions) {
    services(input: $input) {
      nodes {
        id
        name
        isFavorite
      }
    }
  }
`

const useStyles = makeStyles(theme => ({
  addAll: {
    backgroundColor: theme.palette.primary['400'],
  },
  chipContainer: {
    padding: theme.spacing(0.5),
    marginBottom: theme.spacing(2),
    height: '9em',
    overflow: 'auto',
    border: '1px solid #bdbdbd',
  },
  endAdornment: {
    display: 'flex',
    alignItems: 'center',
  },
  noticeText: {
    width: '100%',
    textAlign: 'center',
    alignSelf: 'center',
    lineHeight: '9em',
  },
  searchResults: {
    flexGrow: 1,
    width: '100%',
    overflowY: 'auto',
  },
  topContainer: {
    height: '100%',
  },
  field: {
    width: '100%',
  },
}))

export function CreateAlertServiceSelect(props) {
  const { value, onChange, error } = props
  const [search, setSearch] = useState('')
  const [searchInput, setSearchInput] = useState('')

  const { data, error: queryError, loading } = useQuery(query, {
    variables: {
      input: {
        search,
        favoritesFirst: true,
        omit: value,
        first: 15,
      },
    },
  })

  const fieldRef = useRef()
  const classes = useStyles()
  const searchResults = _.get(data, 'services.nodes', [])

  const queryErrorMsg = allErrors(queryError)
    .map(e => e.message)
    .join('\n')

  let placeholderMsg = null
  if (queryErrorMsg) placeholderMsg = null
  else if (loading || search !== searchInput) placeholderMsg = 'Loading...'
  else if (searchResults.length === 0) placeholderMsg = 'No services found'

  // If the page search param changes, we update state directly.
  useEffect(() => {
    setSearchInput(search)
  }, [search])

  // When typing, we setup a debounce before updating the URL.
  useEffect(() => {
    const t = setTimeout(() => {
      setSearch(searchInput)
    }, DEBOUNCE_DELAY)

    return () => clearTimeout(t)
  }, [searchInput])

  const { labelKey, labelValue } = getServiceLabel(searchInput)

  const addAll = e => {
    e.stopPropagation()
    e.preventDefault()
    const resultIDs = searchResults.map(s => s.id)

    props.onChange(
      _.uniq([...value, ...resultIDs]).slice(0, CREATE_ALERT_LIMIT),
    )
  }

  const selectedServiceChips = value.map(id => {
    return (
      <ServiceChip
        key={id}
        clickable={false}
        id={id}
        style={{ margin: 3 }}
        onClick={e => e.preventDefault()}
        onDelete={() => props.onChange(value.filter(v => v !== id))}
      />
    )
  })

  const notice = (
    <Typography variant='body1' component='p' className={classes.noticeText}>
      Select services using the search box below
    </Typography>
  )

  return (
    <Box
      display='flex'
      justifyContent='flex-start'
      flexDirection='column'
      height='100%'
    >
      <FormControl fullWidth error={Boolean(props.error)}>
        <FormLabel>
          {`Selected Services (${value.length})`}
          {value.length >= CREATE_ALERT_LIMIT && ' - Maximum number allowed'}
        </FormLabel>
        <Paper
          className={classes.chipContainer}
          elevation={0}
          data-cy='service-chip-container'
        >
          {value.length > 0 ? selectedServiceChips : notice}
        </Paper>
        {error && (
          <FormHelperText>
            {(props.error &&
              props.error.message.replace(/^./, str => str.toUpperCase())) ||
              props.hint}
          </FormHelperText>
        )}
      </FormControl>

      <TextField
        fullWidth
        label='Search'
        name='search'
        value={searchInput}
        onChange={e => setSearchInput(e.target.value)}
        InputProps={{
          ref: fieldRef,
          startAdornment: (
            <InputAdornment position='start'>
              <SearchIcon color='action' />
            </InputAdornment>
          ),
          endAdornment: (
            <span className={classes.endAdornment}>
              {searchResults.length > 0 && value.length < CREATE_ALERT_LIMIT && (
                <Chip
                  className={classes.addAll}
                  color='primary' // for white text
                  component='button'
                  label='Add All'
                  size='small'
                  icon={<AddIcon fontSize='small' />}
                  onClick={addAll}
                />
              )}
              <ServiceLabelFilterContainer
                value={{ labelKey, labelValue }}
                onChange={({ labelKey, labelValue }) =>
                  setSearch(labelKey ? `${labelKey}=${labelValue}` : '')
                }
                onReset={() => setSearch('')}
                anchorRef={fieldRef}
              />
            </span>
          ),
        }}
      />
      <Box flexGrow={1} minHeight={0}>
        <Box overflow='auto' flex={1}>
          <List aria-label='select service options'>
            {!!queryErrorMsg && (
              <ListItem>
                <Typography color='error'>{queryErrorMsg}</Typography>
              </ListItem>
            )}

            {searchResults.map(service => (
              <ListItem
                button
                key={service.id}
                disabled={value.length >= CREATE_ALERT_LIMIT}
                onClick={() => onChange([...value, service.id])}
              >
                <ListItemText primary={service.name} />
                {service.isFavorite && (
                  <ListItemIcon>
                    <FavoriteIcon />
                  </ListItemIcon>
                )}
              </ListItem>
            ))}

            {!!placeholderMsg && (
              <ListItem>
                <ListItemText secondary={placeholderMsg} />
              </ListItem>
            )}
          </List>
        </Box>
      </Box>
    </Box>
  )
}
