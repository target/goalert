import React, { useRef, useState, useEffect, MouseEvent } from 'react'
import { gql, useQuery } from '@apollo/client'
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
  Theme,
} from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import ServiceLabelFilterContainer from '../../../services/ServiceFilterContainer'
import { Search as SearchIcon } from '@mui/icons-material'
import { FavoriteIcon } from '../../../util/SetFavoriteButton'
import { ServiceChip } from '../../../util/ServiceChip'
import AddIcon from '@mui/icons-material/Add'
import _ from 'lodash'
import getServiceFilters from '../../../util/getServiceFilters'
import { CREATE_ALERT_LIMIT, DEBOUNCE_DELAY } from '../../../config'

import { allErrors } from '../../../util/errutil'
import { Service } from '../../../../schema'

const query = gql`
  query ($input: ServiceSearchOptions) {
    services(input: $input) {
      nodes {
        id
        name
        isFavorite
      }
    }
  }
`

const useStyles = makeStyles((theme: Theme) => ({
  addAll: {
    marginRight: '0.25em',
  },
  chipContainer: {
    padding: theme.spacing(0.5),
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
  searchInput: {
    marginTop: theme.spacing(2),
  },
  serviceChip: {
    margin: 3,
  },
}))

export interface CreateAlertServiceSelectProps {
  value: string[]
  onChange: (val: string[]) => void
  error?: Error
}

export function CreateAlertServiceSelect(
  props: CreateAlertServiceSelectProps,
): React.JSX.Element {
  const { value, onChange } = props

  const [searchQueryInput, setSearchQueryInput] = useState('')
  const [searchUserInput, setSearchUserInput] = useState('')

  const {
    data,
    error: queryError,
    loading,
  } = useQuery(query, {
    variables: {
      input: {
        search: searchQueryInput,
        favoritesFirst: true,
        omit: value,
        first: 15,
      },
    },
  })

  const fieldRef = useRef<HTMLElement>(null)
  const classes = useStyles()
  const searchResults = _.get(data, 'services.nodes', []).filter(
    ({ id }: { id: string }) => !value.includes(id),
  )

  const queryErrorMsg = allErrors(queryError)
    .map((e) => e.message)
    .join('\n')

  let placeholderMsg = null
  if (queryErrorMsg) {
    placeholderMsg = null
  } else if ((!data && loading) || searchQueryInput !== searchUserInput) {
    placeholderMsg = 'Loading...'
  } else if (searchResults.length === 0) {
    placeholderMsg = 'No services found'
  }

  // debounce search query as user types
  useEffect(() => {
    const t = setTimeout(() => {
      setSearchQueryInput(searchUserInput)
    }, DEBOUNCE_DELAY)

    return () => clearTimeout(t)
  }, [searchUserInput])

  const { labelKey, labelValue } = getServiceFilters(searchUserInput)

  const addAll = (e: MouseEvent<HTMLButtonElement>): void => {
    e.stopPropagation()
    e.preventDefault()
    const resultIDs = searchResults.map((s: { id: string }) => s.id)

    props.onChange(
      _.uniq([...value, ...resultIDs]).slice(0, CREATE_ALERT_LIMIT),
    )
  }

  const selectedServiceChips = value.map((id) => {
    return (
      <ServiceChip
        key={id}
        clickable={false}
        id={id}
        className={classes.serviceChip}
        onClick={(e) => e.preventDefault()}
        onDelete={() => props.onChange(value.filter((v) => v !== id))}
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
      <FormLabel classes={{ root: ' MuiInputLabel-shrink' }}>
        {`Selected Services (${value.length})`}
        {value.length >= CREATE_ALERT_LIMIT && ' - Maximum number allowed'}
      </FormLabel>
      <FormControl fullWidth error={Boolean(props.error)}>
        <Paper
          className={classes.chipContainer}
          elevation={0}
          data-cy='service-chip-container'
        >
          {value.length > 0 ? selectedServiceChips : notice}
        </Paper>
        {Boolean(props.error) && (
          <FormHelperText>
            {props.error?.message.replace(/^./, (str) => str.toUpperCase())}
          </FormHelperText>
        )}
      </FormControl>

      <TextField
        fullWidth
        label='Search'
        name='serviceSearch'
        value={searchUserInput}
        className={classes.searchInput}
        onChange={(e) => setSearchUserInput(e.target.value)}
        InputProps={{
          ref: fieldRef,
          startAdornment: (
            <InputAdornment position='start'>
              <SearchIcon color='action' />
            </InputAdornment>
          ),
          endAdornment: (
            <span className={classes.endAdornment}>
              {searchResults.length > 0 &&
                value.length < CREATE_ALERT_LIMIT && (
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
                  setSearchUserInput(
                    labelKey ? `${labelKey}=${labelValue}` : '',
                  )
                }
                onReset={() => setSearchUserInput('')}
                anchorRef={fieldRef}
              />
            </span>
          ),
        }}
      />
      <Box flexGrow={1} minHeight={0}>
        <Box overflow='auto' flex={1}>
          <List data-cy='service-select' aria-label='select service options'>
            {Boolean(queryErrorMsg) && (
              <ListItem>
                <Typography color='error'>{queryErrorMsg}</Typography>
              </ListItem>
            )}
            {searchResults.map((service: Service) => (
              <ListItem
                button
                data-cy='service-select-item'
                key={service.id}
                disabled={value.length >= CREATE_ALERT_LIMIT}
                onClick={() =>
                  onChange([
                    ...value.filter((id: string) => id !== service.id),
                    service.id,
                  ])
                }
              >
                <ListItemText primary={service.name} />
                {service.isFavorite && (
                  <ListItemIcon>
                    <FavoriteIcon />
                  </ListItemIcon>
                )}
              </ListItem>
            ))}

            {Boolean(placeholderMsg) && (
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
