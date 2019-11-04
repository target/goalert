import React, { useRef } from 'react'
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
  InputLabel,
} from '@material-ui/core'
import { makeStyles, emphasize } from '@material-ui/core/styles'
import { FormField } from '../../../forms'
import ServiceLabelFilterContainer from '../../../services/ServiceLabelFilterContainer'
import { Search as SearchIcon } from '@material-ui/icons'
import FavoriteIcon from '@material-ui/icons/Star'
import AddIcon from '@material-ui/icons/Add'
import OpenInNewIcon from '@material-ui/icons/OpenInNew'
import { ServiceChip } from '../../../util/Chips'
import gql from 'graphql-tag'
import { useQuery } from 'react-apollo'
import _ from 'lodash-es'
import AlertListItem from '../AlertListItem'
import Step0 from './Step0'

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
    margin: 0,
    marginBottom: theme.spacing(2),
    maxHeight: '10em',
    overflow: 'auto',
    border: '1px solid #bdbdbd',
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
  endAdornment: {
    display: 'flex',
    alignItems: 'center',
  },
  noticeBox: {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    height: 150,
  },
  spaceBetween: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  nudgeRight: {
    marginLeft: theme.spacing(1.3),
  },
}))

export default props => {
  const classes = useStyles()
  const fieldRef = useRef()

  const { formFields, mutationStatus } = props

  const labelKey = formFields.searchQuery.split(/(!=|=)/)[0]
  const labelValue = formFields.searchQuery
    .split(/(!=|=)/)
    .slice(2)
    .join('')

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

  const queriedServices = _.get(data, 'services.nodes', [])

  const AddAll = () => (
    <Chip
      component='button'
      label='Add All'
      icon={<AddIcon fontSize='small' />}
      onClick={() => {
        const toAdd = queriedServices.map(s => s.id)
        const newState = formFields.selectedServices.concat(toAdd)
        props.onChange({ selectedServices: newState })
      }}
      className={classes.addAll}
    />
  )

  const OpenAll = () => (
    <Chip
      component='button'
      label='Open All'
      icon={<OpenInNewIcon fontSize='small' />}
      onClick={() => {
        formFields.selectedServices.forEach(id => {
          window.open(`/alerts/${id}`)
        })
      }}
      className={classes.addAll}
    />
  )

  switch (props.activeStep) {
    case 0:
      return <Step0 />
    case 1:
      return (
        <Grid item xs={12}>
          {formFields.selectedServices.length > 0 && (
            <span>
              <InputLabel
                shrink
              >{`Selected Services (${formFields.selectedServices.length})`}</InputLabel>
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
            </span>
          )}
          <FormField
            fullWidth
            label='Search'
            name='searchQuery'
            fieldName='searchQuery'
            component={TextField}
            InputProps={{
              ref: fieldRef,
              startAdornment: (
                <InputAdornment position='start'>
                  <SearchIcon color='action' />
                </InputAdornment>
              ),
              endAdornment: (
                <span className={classes.endAdornment}>
                  {queriedServices.length > 0 && <AddAll />}
                  <ServiceLabelFilterContainer
                    value={{ labelKey, labelValue }}
                    onChange={({ labelKey, labelValue }) =>
                      props.onChange({
                        searchQuery: labelKey
                          ? `${labelKey}=${labelValue}`
                          : '',
                      })
                    }
                    onReset={() =>
                      props.onChange({
                        searchQuery: '',
                      })
                    }
                    anchorRef={fieldRef}
                  />
                </span>
              ),
            }}
          />
          {queriedServices.length > 0 ? (
            <List aria-label='select service options'>
              {queriedServices.map((service, key) => (
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
          ) : (
            <div className={classes.noticeBox}>
              <Typography variant='body1' component='p'>
                {formFields.searchQuery
                  ? 'No services found'
                  : 'Use the search box to select your service(s)'}
              </Typography>
            </div>
          )}
        </Grid>
      )

    case 2:
      return (
        <Paper elevation={0}>
          <Typography variant='subtitle1' component='h3'>
            Summary
          </Typography>
          <Typography
            variant='body1'
            component='p'
            className={classes.nudgeRight}
          >
            {formFields.summary}
          </Typography>
          <Typography variant='subtitle1' component='h3'>
            Details
          </Typography>
          <Typography
            variant='body1'
            component='p'
            className={classes.nudgeRight}
          >
            {formFields.details}
          </Typography>
          <Typography variant='subtitle1' component='h3'>
            {`Selected Services (${formFields.selectedServices.length})`}
          </Typography>

          {formFields.selectedServices.length > 0 && (
            <span>
              <Paper elevation={0}>
                {formFields.selectedServices.map((id, key) => (
                  <ServiceChip
                    key={key}
                    clickable={false}
                    id={id}
                    style={{ margin: 3 }}
                    onClick={e => e.preventDefault()}
                  />
                ))}
              </Paper>
            </span>
          )}
        </Paper>
      )
    case 3:
      const alertsCreated = mutationStatus.alertsCreated || {}
      const graphQLErrors = _.get(
        mutationStatus,
        'alertsFailed.graphQLErrors',
        [],
      )

      const numCreated = Object.keys(alertsCreated).length

      return (
        <Paper elevation={0}>
          {numCreated > 0 && (
            <div>
              <span className={classes.spaceBetween}>
                <Typography variant='subtitle1' component='h3'>
                  {`Successfully created ${numCreated} alerts`}
                </Typography>
                <OpenAll />
              </span>
              <List aria-label='Successfully created alerts'>
                {Object.keys(alertsCreated).map((alias, i) => (
                  <AlertListItem key={i} id={alertsCreated[alias].id} />
                ))}
              </List>
            </div>
          )}

          {graphQLErrors.length > 0 && (
            <div>
              <Typography variant='h6' component='h3'>
                Failed to create alerts on these services:
              </Typography>

              <List aria-label='Failed alerts'>
                {graphQLErrors.map((err, i) => {
                  const index = err.path[0].split(/(\d+)$/)[1]
                  const serviceId = formFields.selectedServices[index]
                  return (
                    <ListItem key={i}>
                      <ServiceChip id={serviceId} />
                    </ListItem>
                  )
                })}
              </List>
            </div>
          )}
        </Paper>
      )
    default:
      return 'Unknown step'
  }
}
