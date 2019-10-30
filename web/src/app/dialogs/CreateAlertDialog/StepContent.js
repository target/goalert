import React from 'react'
import {
  Grid,
  TextField,
  InputAdornment,
  List,
  ListItem,
  ListItemText,
  ListItemIcon,
  Paper,
  Chip,
  Typography,
} from '@material-ui/core'
import { makeStyles, emphasize } from '@material-ui/core/styles'
import { FormField } from '../../forms'
import ServiceLabelFilterContainer from '../../services/ServiceLabelFilterContainer'
import { Search as SearchIcon } from '@material-ui/icons'
import FavoriteIcon from '@material-ui/icons/Star'
import AddIcon from '@material-ui/icons/Add'
import { ServiceChip } from '../../util/Chips'
import gql from 'graphql-tag'
import { useQuery } from 'react-apollo'
import _ from 'lodash-es'

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
  chipContainer: {
    display: 'flex',
    flexWrap: 'wrap',
    padding: theme.spacing(0.5),
    margin: '10px 0px',
  },
  addAll: {
    backgroundColor: theme.palette.grey[100],
    height: theme.spacing(3),
    color: theme.palette.grey[800],
    fontWeight: theme.typography.fontWeightRegular,
    '&:hover, &:focus': {
      backgroundColor: theme.palette.grey[300],
      textDecoration: 'none',
    },
    '&:active': {
      boxShadow: theme.shadows[1],
      backgroundColor: emphasize(theme.palette.grey[300], 0.12),
      textDecoration: 'none',
    },
  },
  addAllWrapper: {
    display: 'flex',
    justifyContent: 'flex-end',
    padding: 3,
  },
}))

export default props => {
  const classes = useStyles()

  const { formFields } = props

  const getLabelKey = () => {
    return formFields.searchQuery.split(/(!=|=)/)[0]
  }

  const getLabelValue = () => {
    return formFields.searchQuery
      .split(/(!=|=)/)
      .slice(2)
      .join('')
  }

  const setLabelKey = newKey => {
    if (newKey) {
      props.onChange({ searchQuery: newKey + '=' })
    } else {
      props.onChange({ searchQuery: '' }) // clear search if clearing key
    }
  }

  const setLabelValue = newValue => {
    // should be disabled if empty, but just in case :)
    if (!getLabelKey()) return

    if (newValue) {
      props.onChange({ searchQuery: getLabelKey() + '=' + newValue })
    } else {
      props.onChange({ searchQuery: getLabelKey() + '=' })
    }
  }

  // TODO loading error handles
  const { data } = useQuery(query, {
    variables: {
      input: {
        search: formFields.searchQuery,
        favoritesFirst: true,
        omit: formFields.selectedServices,
      },
    },
    skip: formFields.searchQuery.length === 0,
  })

  const AddAll = () => (
    <div className={classes.addAllWrapper}>
      <Chip
        component='button'
        label='Add All'
        icon={<AddIcon fontSize='small' />}
        onClick={() => {
          const toAdd = _.get(data, 'services.nodes', []).map(s => s.id)
          const newState = formFields.selectedServices.concat(toAdd)
          props.onChange({ selectedServices: newState })
        }}
        className={classes.addAll}
      />
    </div>
  )

  switch (props.activeStep) {
    case 0:
      return (
        <Grid container spacing={2}>
          <Grid item xs={12}>
            <FormField
              fullWidth
              label='Alert Summary'
              name='summary'
              required
              component={TextField}
            />
          </Grid>
          <Grid item xs={12}>
            <FormField
              fullWidth
              label='Alert Details'
              name='details'
              required
              component={TextField}
            />
          </Grid>
        </Grid>
      )
    case 1:
      return (
        <Grid item xs={12}>
          {formFields.selectedServices.length > 0 && (
            <Paper className={classes.chipContainer} elevation={0}>
              {formFields.selectedServices.map((id, key) => {
                return (
                  <ServiceChip
                    key={key}
                    clickable={false}
                    id={id}
                    style={{ margin: 3 }}
                    onClick={e => e.preventDefault()}
                    onDelete={() =>
                      props.onChange({
                        selectedServices: formFields.selectedServices.filter(
                          sid => sid !== id,
                        ),
                      })
                    }
                  />
                )
              })}
            </Paper>
          )}
          <FormField
            fullWidth
            label='Search Query'
            name='searchQuery'
            fieldName='searchQuery'
            required
            component={TextField}
            InputProps={{
              startAdornment: (
                <InputAdornment position='start'>
                  <SearchIcon color='action' />
                </InputAdornment>
              ),
              endAdornment: (
                <ServiceLabelFilterContainer
                  c={console.log(formFields)}
                  labelKey={getLabelKey()}
                  labelValue={getLabelValue()}
                  onKeyChange={setLabelKey}
                  onValueChange={setLabelValue}
                  onReset={() =>
                    props.onChange({
                      searchQuery: '',
                    })
                  }
                />
              ),
            }}
          />
          {_.get(data, 'services.nodes', []).length > 0 && <AddAll />}
          {formFields.searchQuery && (
            <List aria-label='select service options'>
              {_.get(data, 'services.nodes', []).map((service, key) => (
                <ListItem
                  button
                  key={key}
                  disabled={formFields.selectedServices.indexOf(service) !== -1}
                  onClick={() => {
                    const newState = [
                      ...formFields.selectedServices,
                      service.id,
                    ]
                    props.onChange({ selectedServices: newState })
                  }}
                >
                  <ListItemText primary={service.name} />
                  {service.isFavorite && (
                    <ListItemIcon>
                      <FavoriteIcon />
                    </ListItemIcon>
                  )}
                </ListItem>
              ))}
            </List>
          )}
        </Grid>
      )

    case 2:
      return (
        <Paper elevation={0}>
          <Typography variant='h6' component='h3'>
            Summary: {formFields.summary}
          </Typography>
          <Typography variant='h6' component='h3'>
            Details: {formFields.details}
          </Typography>
          {formFields.selectedServices.map((id, key) => {
            return (
              <ServiceChip
                key={key}
                clickable={false}
                id={id}
                style={{ margin: 3 }}
                onClick={e => e.preventDefault()}
              />
            )
          })}
        </Paper>
      )
    case 3:
      return 'revue ur alerts'
    default:
      return 'Unknown step'
  }
}
