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
} from '@material-ui/core'
import { makeStyles } from '@material-ui/core/styles'

import { FormField } from '../../forms'
import ServiceLabelFilterContainer from '../../services/ServiceLabelFilterContainer'
import { Search as SearchIcon } from '@material-ui/icons'
import FavoriteIcon from '@material-ui/icons/Star'
import { ServiceChip } from '../../util/Chips'

const useStyles = makeStyles(theme => ({
  chipContainer: {
    display: 'flex',
    flexWrap: 'wrap',
    padding: theme.spacing(0.5),
    margin: '10px 0px',
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
            <Paper className={classes.chipContainer}>
              {formFields.selectedServices.map((service, key) => {
                return (
                  <ServiceChip
                    key={key}
                    clickable={false}
                    id={service.id}
                    name={service.name}
                    style={{ margin: 3 }}
                    onClick={e => e.preventDefault()}
                    onDelete={() =>
                      props.onChange({
                        selectedServices: formFields.selectedServices.filter(
                          s => s.id !== service.id,
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
          {formFields.searchQuery && (
            <List aria-label='select service options'>
              {formFields.services.map((service, key) => (
                <ListItem
                  button
                  key={key}
                  disabled={formFields.selectedServices.indexOf(service) !== -1}
                  onClick={() => {
                    const newState = [...formFields.selectedServices, service]
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
      return 'plz confirm ur info'
    default:
      return 'Unknown step'
  }
}
