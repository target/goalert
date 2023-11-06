import React, { ReactNode } from 'react'
import {
  Typography,
  Grid,
  Divider,
  TextField,
  MenuItem,
  Fab,
  FormControlLabel,
  Switch,
} from '@mui/material'
import { FormContainer, FormField } from '../../forms'
import {
  Content,
  IntegrationKey,
  ServiceRuleActionInput,
  ServiceRuleFilterInput,
  ServiceRuleFilterValueType,
} from '../../../schema'
import makeStyles from '@mui/styles/makeStyles'
import AddIcon from '@mui/icons-material/Add'
import ClearIcon from '@mui/icons-material/Clear'
import { FieldError } from '../../util/errutil'
import MaterialSelect from '../../selection/MaterialSelect'
import { SlackChannelSelect } from '../../selection'
import toTitleCase from '../../util/toTitleCase'

export interface ServiceRuleValue {
  id?: string
  name: string
  serviceID?: string
  filters: ServiceRuleFilterValue[]
  sendAlert: boolean
  actions: ServiceRuleActionValue[]
  integrationKeys: IntegrationKeySelectVal[]

  customFields?: CustomFields
}

export interface CustomFields {
  summary: string
  details: string
}

interface ServiceRuleFilterValue {
  field: string
  operator: string
  value: string
  valueType: ServiceRuleFilterValueType
}

interface ServiceRuleActionValue {
  destType: string
  contents: ContentValue[]
}

interface ContentValue {
  prop: string
  value: string
}

interface IntegrationKeySelectVal {
  label: string
  value: string
}

interface ServiceRuleFormProps {
  value: ServiceRuleValue
  serviceID: string
  errors: FieldError[]

  actionsError: boolean

  onChange: (val: ServiceRuleValue) => void
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

export const destType = {
  ALERT: 'ALERT',
  SLACK: 'SLACK',
  SERVICENOW: 'SERVICENOW',
  WEBHOOK: 'WEBHOOK',
  EMAIL: 'EMAIL',
}

const destinations = [
  { label: 'Slack', value: destType.SLACK },
  { label: 'ServiceNow Proactice Incident', value: destType.SERVICENOW },
  { label: 'Webhook', value: destType.WEBHOOK },
  { label: 'Email', value: destType.EMAIL },
]

const useStyles = makeStyles({
  itemContent: {
    marginTop: '0.5em',
  },
  itemTitle: {
    paddingBottom: 0,
  },
  fab: {
    marginTop: '0.5em',
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
      customFields: formProps.value.customFields,
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
          contents: [],
        },
      ],
      integrationKeys: formProps.value.integrationKeys,
      customFields: formProps.value.customFields,
    })
  }

  const handleActionDestSelect = (dest: string, actionIdx: number): void => {
    const actions = formProps.value.actions
    actions[actionIdx].destType = dest
    switch (dest) {
      case destType.SLACK:
        actions[actionIdx].contents = [
          { prop: 'channel_id', value: '' },
          { prop: 'channel', value: '' },
          { prop: 'message', value: '' },
        ]
        break
      case destType.SERVICENOW:
        actions[actionIdx].contents = [
          { prop: 'assignment_group', value: '' },
          { prop: 'caused_by_this', value: '' },
          { prop: 'short_description', value: '' },
        ]
        break
      case destType.WEBHOOK:
        actions[actionIdx].contents = [
          { prop: 'body', value: '' },
          { prop: 'URL', value: '' },
          { prop: 'Method', value: '' },
        ]
        break
      case destType.EMAIL:
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
      customFields: formProps.value.customFields,
    })
  }

  const handleAddCustomAlertFields = (e: boolean): void => {
    if (e) {
      formProps.onChange({
        name: formProps.value.name,
        serviceID,
        filters: formProps.value.filters,
        sendAlert: formProps.value.sendAlert,
        actions: formProps.value.actions,
        integrationKeys: formProps.value.integrationKeys,
        customFields: {
          summary: '',
          details: '',
        },
      })
    } else {
      formProps.onChange({
        name: formProps.value.name,
        serviceID,
        filters: formProps.value.filters,
        sendAlert: formProps.value.sendAlert,
        actions: formProps.value.actions,
        integrationKeys: formProps.value.integrationKeys,
        customFields: undefined,
      })
    }
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
      customFields: formProps.value.customFields,
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
      customFields: formProps.value.customFields,
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
      customFields: formProps.value.customFields,
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
        <Grid item>
          <FormControlLabel
            label='Create Alert'
            labelPlacement='end'
            control={
              <FormField noError component={Switch} checkbox name='sendAlert' />
            }
          />
          <Typography variant='body2'>
            If enabled, all signals matching this rule will create a standard
            GoAlert alert and trigger the escalation policy steps for this
            service.
          </Typography>
          <FormControlLabel
            label='Custom Fields'
            labelPlacement='end'
            control={
              <Switch
                checked={
                  formProps.value.customFields !== undefined &&
                  formProps.value.sendAlert
                }
                onChange={(e) => handleAddCustomAlertFields(e.target.checked)}
              />
            }
            disabled={!formProps.value.sendAlert}
          />
        </Grid>
        {formProps.value.customFields && formProps.value.sendAlert && (
          <React.Fragment>
            <Grid item style={{ flexGrow: 1 }} xs={12}>
              <FormField
                fullWidth
                component={TextField}
                label='Summary'
                name='customFields.summary'
                required
              />
            </Grid>
            <Grid item style={{ flexGrow: 1 }} xs={12}>
              <FormField
                fullWidth
                component={TextField}
                label='Details'
                name='customFields.details'
                required
              />
            </Grid>
          </React.Fragment>
        )}
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
        <Grid item xs={12}>
          <Field label='Destinations'>
            {formProps.value.actions.map(
              (action: ServiceRuleActionInput, actionIdx: number) => {
                if (action.destType !== destType.ALERT) {
                  return (
                    <React.Fragment key={actionIdx}>
                      <Grid container>
                        <Grid item style={{ flexGrow: 1 }} xs={11}>
                          <TextField
                            fullWidth
                            select
                            label='Select Destination'
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

                        <Grid
                          item
                          style={{ flexGrow: 1 }}
                          xs={1}
                          sx={{ pl: 1 }}
                        >
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

                        {action.destType === destType.SLACK && (
                          <Grid
                            item
                            style={{ flexGrow: 1, marginTop: '1em' }}
                            xs={12}
                          >
                            <FormField
                              required
                              component={SlackChannelSelect}
                              fullWidth
                              label='Select Channel(s)'
                              name={`actions[${actionIdx}].contents[0].value`}
                              value={
                                formProps.value.actions[actionIdx].contents[0]
                                  .value
                              }
                            />
                          </Grid>
                        )}
                        {action.contents &&
                          action.contents.map(
                            (c: Content, contentIdx: number) => {
                              if (
                                (action.destType === destType.SLACK &&
                                  c.prop === 'channel') ||
                                (action.destType === destType.SLACK &&
                                  c.prop === 'channel_id')
                              ) {
                                return <React.Fragment key={contentIdx} />
                              }
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
                                    label={toTitleCase(
                                      c.prop.replace(/_/g, ' '),
                                    )}
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
                }
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
