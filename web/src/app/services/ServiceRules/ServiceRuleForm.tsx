import React, { ReactNode } from 'react'
import {
  Typography,
  Grid,
  Divider,
  TextField,
  MenuItem,
  Fab,
  Checkbox,
  FormControlLabel,
} from '@mui/material'
import { FormContainer, FormField } from '../../forms'
import {
  Content,
  CreateServiceRuleInput,
  IntegrationKey,
  ServiceRuleActionInput,
  ServiceRuleFilterInput,
  UpdateServiceRuleInput,
} from '../../../schema'
import makeStyles from '@mui/styles/makeStyles'
import AddIcon from '@mui/icons-material/Add'
import ClearIcon from '@mui/icons-material/Clear'
import { FieldError } from '../../util/errutil'
import MaterialSelect from '../../selection/MaterialSelect'

interface ServiceRuleFormProps {
  value: CreateServiceRuleInput | UpdateServiceRuleInput
  serviceID: string
  errors: FieldError[]

  actionsError: boolean

  onChange: (val: CreateServiceRuleInput | UpdateServiceRuleInput) => void
  disabled: boolean
  integrationKeys: IntegrationKey[]
}

type FieldProps = {
  children: ReactNode
  label: string
}

const operators = [
  '==',
  '!=',
  '<=',
  '>=',
  '<',
  '>',
  'startsWith',
  'endsWith',
  'contains',
]

const destinations = [
  { label: 'Slack', value: 'slack' },
  { label: 'ServiceNow Proactice Incident', value: 'servicenow' },
  { label: 'Webhook', value: 'webhook' },
  { label: 'Email', value: 'email' },
]

const useStyles = makeStyles({
  itemContent: {
    marginTop: '0.5em',
  },
  itemTitle: {
    paddingBottom: 0,
  },
  fab: {
    marginTop: '1em',
  },
  filterRow: {
    marginBottom: '1em',
  },
  actionDivider: {
    marginTop: '1em',
    marginBottom: '1em',
  },
})

function Field(props: FieldProps): JSX.Element {
  const classes = useStyles()
  return (
    <Grid item xs={12}>
      <Typography
        variant='subtitle1'
        component='h3'
        className={classes.itemTitle}
      >
        {props.label}
      </Typography>

      <Divider />

      <div className={classes.itemContent}>{props.children}</div>
    </Grid>
  )
}

export default function ServiceRuleForm(
  props: ServiceRuleFormProps,
): JSX.Element {
  const classes = useStyles()
  const { serviceID, actionsError, integrationKeys, ...formProps } = props

  const handleAddFilter = (): void => {
    formProps.onChange({
      name: formProps.value.name,
      serviceID,
      filters: [
        ...formProps.value.filters,
        {
          field: '',
          operator: '==',
          value: '',
          valueType: 'UNKNOWN',
        },
      ],
      sendAlert: formProps.value.sendAlert,
      actions: formProps.value.actions,
      integrationKeys: formProps.value.integrationKeys,
    })
  }

  const handleAddAction = (): void => {
    formProps.onChange({
      name: formProps.value.name,
      serviceID,
      filters: formProps.value.filters,
      sendAlert: formProps.value.sendAlert,
      actions: [
        ...formProps.value.actions,
        {
          destType: '',
          destID: '',
          destValue: '',
          contents: [],
        },
      ],
      integrationKeys: formProps.value.integrationKeys,
    })
  }

  const handleActionDestSelect = (dest: string, actionIdx: number): void => {
    const actions = formProps.value.actions
    actions[actionIdx].destType = dest
    switch (dest) {
      case 'slack':
        actions[actionIdx].contents = [
          { prop: 'channel', value: '' },
          { prop: 'message', value: '' },
        ]
        break
      case 'servicenow':
        actions[actionIdx].contents = [
          { prop: 'assignment_group', value: '' },
          { prop: 'caused_by_this', value: '' },
          { prop: 'short_description', value: '' },
        ]
        break
      case 'webhook':
        actions[actionIdx].contents = [
          { prop: 'body', value: '' },
          { prop: 'URL', value: '' },
          { prop: 'Method', value: '' },
        ]
        break
      case 'email':
        actions[actionIdx].contents = [
          { prop: 'address', value: '' },
          { prop: 'subject', value: '' },
          { prop: 'body', value: '' },
        ]
        break
      default:
        actions[actionIdx].contents = []
    }

    formProps.onChange({
      name: formProps.value.name,
      serviceID,
      filters: formProps.value.filters,
      sendAlert: formProps.value.sendAlert,
      actions,
      integrationKeys: formProps.value.integrationKeys,
    })
  }

  const handleDeleteFilter = (deleteFilter: ServiceRuleFilterInput): void => {
    formProps.onChange({
      name: formProps.value.name,
      serviceID,
      filters: formProps.value.filters.filter(
        (filter: ServiceRuleFilterInput) => {
          return filter !== deleteFilter
        },
      ),
      sendAlert: formProps.value.sendAlert,
      actions: formProps.value.actions,
      integrationKeys: formProps.value.integrationKeys,
    })
  }

  const handleSelectFilterOperator = (
    filterIdx: number,
    operator: string,
  ): void => {
    const updatedFilters = formProps.value.filters
    updatedFilters[filterIdx].operator = operator
    formProps.onChange({
      name: formProps.value.name,
      serviceID,
      filters: updatedFilters,
      sendAlert: formProps.value.sendAlert,
      actions: formProps.value.actions,
      integrationKeys: formProps.value.integrationKeys,
    })
  }

  const handleDeleteAction = (deleteAction: ServiceRuleActionInput): void => {
    formProps.onChange({
      name: formProps.value.name,
      serviceID,
      filters: formProps.value.filters,
      sendAlert: formProps.value.sendAlert,
      actions: formProps.value.actions.filter(
        (action: ServiceRuleActionInput) => {
          return action !== deleteAction
        },
      ),
      integrationKeys: formProps.value.integrationKeys,
    })
  }

  return (
    <FormContainer {...formProps} optionalLabels>
      <Grid container spacing={2}>
        <Grid item style={{ flexGrow: 1 }} xs={12}>
          <FormField
            fullWidth
            component={TextField}
            label='Name'
            name='name'
            required
          />
        </Grid>
        <Grid item style={{ flexGrow: 1 }} xs={12}>
          <FormField
            value={formProps.value.integrationKeys}
            component={MaterialSelect}
            name='integrationKeys'
            label='Select Integration Key(s)'
            required
            fullWidth
            multiple
            options={integrationKeys.map((key: IntegrationKey) => ({
              label: key.name,
              value: key.id,
            }))}
          />
        </Grid>
        <Grid item xs={12}>
          <Field label='Filters'>
            {formProps.value.filters.map(
              (v: ServiceRuleFilterInput, filterIdx: number) => {
                return (
                  <Grid
                    container
                    key={filterIdx}
                    spacing={1}
                    className={classes.filterRow}
                    style={{ flexGrow: 1 }}
                  >
                    <Grid item xs={4}>
                      <FormField
                        fullWidth
                        component={TextField}
                        label='Field'
                        name={`filters[${filterIdx}].field`}
                        required
                        value={formProps.value.filters[filterIdx].field}
                      />
                    </Grid>
                    <Grid item xs={3}>
                      <TextField
                        required
                        fullWidth
                        select
                        label='Operator'
                        name={`filters[${filterIdx}].operator`}
                        value={formProps.value.filters[filterIdx].operator}
                      >
                        {operators.map((op: string) => (
                          <MenuItem
                            key={op}
                            value={op}
                            onClick={() =>
                              handleSelectFilterOperator(filterIdx, op)
                            }
                          >
                            {op}
                          </MenuItem>
                        ))}
                      </TextField>
                    </Grid>
                    <Grid item xs={4}>
                      <FormField
                        fullWidth
                        component={TextField}
                        label='Value'
                        name={`filters[${filterIdx}].value`}
                        required
                        value={formProps.value.filters[filterIdx].value}
                      />
                    </Grid>
                    <Grid item xs={1}>
                      <Fab
                        className={classes.fab}
                        size='small'
                        color='error'
                        aria-label='delete'
                        onClick={() => handleDeleteFilter(v)}
                      >
                        <ClearIcon />
                      </Fab>
                    </Grid>
                  </Grid>
                )
              },
            )}
          </Field>
          <Fab
            className={classes.fab}
            size='small'
            color='primary'
            aria-label='clear'
            onClick={() => handleAddFilter()}
          >
            <AddIcon />
          </Fab>
        </Grid>
        <Grid item>
          <FormControlLabel
            label='Create Alert'
            labelPlacement='end'
            control={
              <FormField
                noError
                component={Checkbox}
                checkbox
                name='sendAlert'
              />
            }
          />
        </Grid>

        <Grid item xs={12}>
          <Field label='Actions'>
            {formProps.value.actions.map(
              (action: ServiceRuleActionInput, actionIdx: number) => {
                return (
                  <React.Fragment key={actionIdx}>
                    <Grid container>
                      <Grid item style={{ flexGrow: 1 }} xs={11}>
                        <TextField
                          fullWidth
                          select
                          label='Select Action'
                          value={action.destType}
                        >
                          {destinations.map((dest) => (
                            <MenuItem
                              key={dest.value}
                              onClick={() =>
                                handleActionDestSelect(dest.value, actionIdx)
                              }
                              value={dest.value}
                            >
                              {dest.label}
                            </MenuItem>
                          ))}
                        </TextField>
                      </Grid>

                      <Grid item style={{ flexGrow: 1 }} xs={1} sx={{ pl: 1 }}>
                        <Fab
                          className={classes.fab}
                          size='small'
                          color='error'
                          aria-label='delete'
                          onClick={() => handleDeleteAction(action)}
                        >
                          <ClearIcon />
                        </Fab>
                      </Grid>
                      {action.contents &&
                        action.contents.map(
                          (c: Content, contentIdx: number) => {
                            return (
                              <Grid
                                key={c.prop}
                                item
                                style={{ flexGrow: 1, marginTop: '1em' }}
                                xs={12}
                              >
                                <FormField
                                  fullWidth
                                  component={TextField}
                                  label={c.prop}
                                  name={`actions[${actionIdx}].contents[${contentIdx}].value`}
                                  required
                                />
                              </Grid>
                            )
                          },
                        )}
                    </Grid>
                    {formProps.value.actions.length > 1 &&
                      actionIdx !== formProps.value.actions.length - 1 && (
                        <Divider className={classes.actionDivider} />
                      )}
                  </React.Fragment>
                )
              },
            )}
            {actionsError && (
              <Typography color='error'>
                At least 1 action is required.
              </Typography>
            )}
          </Field>
          <Fab
            className={classes.fab}
            size='small'
            color='primary'
            aria-label='add'
            onClick={() => handleAddAction()}
          >
            <AddIcon />
          </Fab>
        </Grid>
      </Grid>
    </FormContainer>
  )
}
