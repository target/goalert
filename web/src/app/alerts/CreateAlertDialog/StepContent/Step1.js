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
  Typography,
  InputLabel,
  Chip,
} from '@material-ui/core'
import { makeStyles } from '@material-ui/core/styles'
import { FormField } from '../../../forms'
import ServiceLabelFilterContainer from '../../../services/ServiceLabelFilterContainer'
import { Search as SearchIcon } from '@material-ui/icons'
import FavoriteIcon from '@material-ui/icons/Star'
import { ServiceChip } from '../../../util/Chips'
import AddIcon from '@material-ui/icons/Add'

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
}))

const CREATE_ALERT_LIMIT = 35

export default function Step1(props) {
  const { formFields, queriedServices } = props

  const fieldRef = useRef()
  const classes = useStyles()

  const labelKey = formFields.searchQuery.split(/(!=|=)/)[0]
  const labelValue = formFields.searchQuery
    .split(/(!=|=)/)
    .slice(2)
    .join('')

  const AddAll = () => (
    <Chip
      className={classes.addAll}
      color='primary' // for white text
      component='button'
      label='Add All'
      size='small'
      icon={<AddIcon fontSize='small' />}
      onClick={() => {
        const toAdd = queriedServices.map(s => s.id)

        // build newState
        let newState = formFields.selectedServices
        toAdd.forEach(s => {
          if (newState.length < CREATE_ALERT_LIMIT) {
            newState = newState.concat(s)
          }
        })

        props.onChange({ selectedServices: newState })
      }}
    />
  )

  const serviceChips = formFields.selectedServices.map(id => {
    return (
      <ServiceChip
        key={id}
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
  })

  const notice = (
    <Typography variant='body1' component='p' className={classes.noticeText}>
      Select services using the search box below
    </Typography>
  )

  return (
    <Grid item xs={12}>
      <InputLabel shrink>
        {`Selected Services (${formFields.selectedServices.length})`}
        {formFields.selectedServices.length === CREATE_ALERT_LIMIT &&
          ' - Maximum number allowed'}
      </InputLabel>
      <Paper
        className={classes.chipContainer}
        elevation={0}
        data-cy='service-chip-container'
      >
        {formFields.selectedServices.length > 0 ? serviceChips : notice}
      </Paper>

      <FormField
        fullWidth
        label='Search'
        name='searchQuery'
        fieldName='searchQuery'
        required
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
              {queriedServices.length > 0 &&
                formFields.selectedServices.length < CREATE_ALERT_LIMIT && (
                  <AddAll />
                )}
              <ServiceLabelFilterContainer
                value={{ labelKey, labelValue }}
                onChange={({ labelKey, labelValue }) =>
                  props.onChange({
                    searchQuery: labelKey ? `${labelKey}=${labelValue}` : '',
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

      <List aria-label='select service options'>
        {queriedServices.map((service, key) => (
          <ListItem
            button
            key={key}
            disabled={formFields.selectedServices.length >= CREATE_ALERT_LIMIT}
            onClick={() => {
              const newState = [...formFields.selectedServices, service.id]
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

        {queriedServices.length === 0 && (
          <ListItem>
            <ListItemText secondary='No services found' />
          </ListItem>
        )}
      </List>
    </Grid>
  )
}
