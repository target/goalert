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

import { FormContainer, FormField } from '../../forms'
import ServiceLabelFilterContainer from '../../services/ServiceLabelFilterContainer'
import { Search as SearchIcon } from '@material-ui/icons'
import FavoriteIcon from '@material-ui/icons/Star'
import { ServiceChip } from '../../util/Chips'

const useStyles = makeStyles(theme => ({
  root: {
    display: 'flex',
    justifyContent: 'center',
    flexWrap: 'wrap',
    padding: theme.spacing(0.5),
  },
  chip: {
    margin: theme.spacing(0.5),
  },
}))

export default props => {
  const classes = useStyles()

  const { formFields } = props

  switch (props.activeStep) {
    case 0:
      return (
        <Grid container spacing={2}>
          <FormContainer onChange={props.onChange}>
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
          </FormContainer>
        </Grid>
      )
    case 1:
      return (
        <Grid item xs={12}>
          <Paper className={classes.root}>
            {formFields.selectedServices.map(service => {
              return (
                <ServiceChip
                  id={service.id}
                  name={service.name}
                  // className={classes.chip}
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
          <FormContainer onChange={props.onChange}>
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
                    labelKey={formFields.labelKey}
                    labelValue={formFields.labelValue}
                    onKeyChange={newKey => {
                      newKey = `${newKey}=`
                      const newState = { labelKey: newKey, searchQuery: newKey }
                      props.setFormFields(prevState => ({
                        ...prevState,
                        ...newState,
                      }))
                    }}
                    onValueChange={newValue => {
                      let newState = { labelValue: newValue }
                      if (formFields.searchQuery.endsWith('=')) {
                        newState['searchQuery'] =
                          formFields.searchQuery + newValue
                      }
                      props.setFormFields(prevState => ({
                        ...prevState,
                        ...newState,
                      }))
                    }}
                  />
                ),
              }}
            />
            {formFields.searchQuery && (
              <List component='nav' aria-label='main mailbox folders'>
                {formFields.services.map(service => (
                  <ListItem
                    button
                    onClick={() => {
                      if (formFields.selectedServices.indexOf(service) !== -1) {
                        return
                      }
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
          </FormContainer>
        </Grid>
      )

    case 2:
      return 'plz confirm ur info'
    default:
      return 'Unknown step'
  }
}
