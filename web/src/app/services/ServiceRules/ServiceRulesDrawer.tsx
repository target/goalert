import React, { useState } from 'react'
import {
  ClickAwayListener,
  Divider,
  Drawer,
  Grid,
  List,
  ListItem,
  ListItemText,
  Toolbar,
  Typography,
  Button,
  ButtonGroup,
  Chip,
} from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import {
  Content,
  IntegrationKey,
  ServiceRule,
  ServiceRuleAction,
  ServiceRuleFilter,
} from '../../../schema'
import ServiceRuleDeleteDialog from './ServiceRuleDeleteDialog'
import ServiceRuleEditDialog, { getCustomFields } from './ServiceRuleEditDialog'
import { destType } from './ServiceRuleForm'
import toTitleCase from '../../util/toTitleCase'

interface ServiceRulesDrawerProps {
  onClose: () => void
  integrationKeys: IntegrationKey[]
  rule: ServiceRule | null
}

const useStyles = makeStyles(() => ({
  buttons: {
    textAlign: 'right',
    width: '30vw',
    padding: '15px 10px',
  },
  chip: {
    margin: '0.5em',
  },
}))

export default function ServiceRulesDrawer(
  props: ServiceRulesDrawerProps,
): JSX.Element {
  const { onClose, rule, integrationKeys } = props
  const classes = useStyles()
  const isOpen = Boolean(rule?.id)
  const [editRule, setEditRule] = useState<ServiceRule | null>(null)
  const [deleteRule, setDeleteRule] = useState<string | null>(null)
  const customFields = getCustomFields(rule)

  return (
    <ClickAwayListener onClickAway={onClose} mouseEvent='onMouseUp'>
      <Drawer
        anchor='right'
        open={isOpen}
        variant='persistent'
        data-cy='service-rule-details-drawer'
      >
        <Toolbar />
        {deleteRule ? (
          <ServiceRuleDeleteDialog
            onClose={(): void => {
              setDeleteRule(null)
              onClose()
            }}
            ruleID={deleteRule}
          />
        ) : null}
        {editRule ? (
          <ServiceRuleEditDialog
            serviceID={rule?.serviceID || ''}
            rule={editRule}
            onClose={(): void => {
              setEditRule(null)
              onClose()
            }}
            integrationKeys={integrationKeys}
          />
        ) : null}
        <Grid style={{ width: '30vw' }}>
          <Typography variant='h6' style={{ margin: '16px' }}>
            Signal Rule Details
          </Typography>
          <Divider />
          <List disablePadding>
            <ListItem divider>
              <ListItemText primary='Name' secondary={rule?.name} />
            </ListItem>
          </List>
          <List disablePadding>
            <ListItem divider>
              <ListItemText
                primary='Integration Keys'
                secondary={rule?.integrationKeys.map((key: IntegrationKey) => (
                  <Chip
                    key={key.id}
                    label={key.name}
                    className={classes.chip}
                  />
                ))}
              />
            </ListItem>
          </List>
          <List disablePadding>
            <ListItem>
              <ListItemText
                primary='Create Alert'
                secondary={
                  <React.Fragment>
                    <Typography>
                      {rule?.sendAlert ? 'True' : 'False'}
                    </Typography>
                    {customFields && (
                      <List disablePadding>
                        <ListItem>
                          <ListItemText
                            primary='Custom Summary'
                            secondary={customFields.summary}
                          />
                        </ListItem>
                        <ListItem>
                          <ListItemText
                            primary='Custom Details'
                            secondary={customFields.details}
                          />
                        </ListItem>
                      </List>
                    )}
                  </React.Fragment>
                }
              />
            </ListItem>
          </List>
          <Divider />
          {rule && rule.filters.length > 0 && (
            <List disablePadding>
              <ListItem divider>
                <ListItemText
                  primary='Filters'
                  secondary={rule?.filters.map(
                    (f: ServiceRuleFilter, idx: number) => (
                      <Chip
                        key={idx}
                        label={`${f.field} ${f.operator} ${f.value}`}
                        className={classes.chip}
                      />
                    ),
                  )}
                />
              </ListItem>
            </List>
          )}
          <Divider />
          <List disablePadding>
            <ListItem divider>
              <ListItemText
                primary='Destinations'
                secondary={rule?.actions.map(
                  (action: ServiceRuleAction, idx: number) => {
                    if (action.destType !== destType.ALERT) {
                      return (
                        <List disablePadding key={idx}>
                          <ListItem>
                            <ListItemText
                              primary={action.destType}
                              secondary={action.contents.map(
                                (content: Content, idx: number) => (
                                  <List disablePadding key={idx}>
                                    <ListItem>
                                      <ListItemText
                                        primary={toTitleCase(
                                          content.prop.replace(/_/g, ' '),
                                        )}
                                        secondary={content.value}
                                      />
                                    </ListItem>
                                  </List>
                                ),
                              )}
                            />
                          </ListItem>
                        </List>
                      )
                    }
                  },
                )}
              />
            </ListItem>
          </List>
          <Grid className={classes.buttons}>
            <ButtonGroup variant='contained'>
              <Button onClick={() => setDeleteRule(rule?.id || '')}>
                Delete
              </Button>
              <Button onClick={() => setEditRule(rule)}>Edit</Button>
            </ButtonGroup>
          </Grid>
        </Grid>
      </Drawer>
    </ClickAwayListener>
  )
}
