import React from 'react'
import {
  Grid,
  TextField,
  InputAdornment,
  List,
  ListItem,
  ListItemText,
  ListItemIcon,
} from '@material-ui/core'
import { FormContainer, FormField } from '../../forms'
import ServiceLabelFilterContainer from '../../services/ServiceLabelFilterContainer'
import { Search as SearchIcon } from '@material-ui/icons'
import FavoriteIcon from '@material-ui/icons/Star'

export default props => {
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
                    labelKey={props.formFields.labelKey}
                    labelValue={props.formFields.labelValue}
                    onKeyChange={newKey => {
                      newKey = `${newKey}=`
                      const newState = { labelKey: newKey, searchQuery: newKey }
                      props.setFormFields(prevState => ({
                        ...prevState,
                        ...newState,
                      }))
                    }}
                    onValueChange={newValue => {
                      const newState = { labelValue: newValue }
                      props.setFormFields(prevState => ({
                        ...prevState,
                        ...newState,
                      }))
                    }}
                  />
                ),
              }}
            />
            {props.formFields.searchQuery && (
              <List component='nav' aria-label='main mailbox folders'>
                {props.formFields.services.map(service => (
                  <ListItem button>
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
