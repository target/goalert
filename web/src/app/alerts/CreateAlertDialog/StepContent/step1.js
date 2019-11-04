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
import { makeStyles, emphasize } from '@material-ui/core/styles'
import { FormField } from '../../../forms'
import ServiceLabelFilterContainer from '../../../services/ServiceLabelFilterContainer'
import { Search as SearchIcon } from '@material-ui/icons'
import FavoriteIcon from '@material-ui/icons/Star'
import { ServiceChip } from '../../../util/Chips'
import AddIcon from '@material-ui/icons/Add'

const useStyles = makeStyles(theme => ({
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
}))

export default props => {
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
      {queriedServices.length > 0 ? (
        <List aria-label='select service options'>
          {queriedServices.map((service, key) => (
            <ListItem
              button
              key={key}
              disabled={formFields.selectedServices.indexOf(service) !== -1}
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
}
