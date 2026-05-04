import React, { useState, ReactElement, Suspense } from 'react'
import { useQuery, gql } from 'urql'
import Button from '@mui/material/Button'
import Grid from '@mui/material/Grid'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import Typography from '@mui/material/Typography'
import Chip from '@mui/material/Chip'
import CreateFAB from '../lists/CreateFAB'
import ServiceIMAPConfigDialog from './IMAP/ServiceIMAPConfigDialog'
import IMAPFilterRuleCreateDialog from './IMAP/IMAPFilterRuleCreateDialog'
import IMAPFilterRuleEditDialog from './IMAP/IMAPFilterRuleEditDialog'
import IMAPFilterRuleDeleteDialog from './IMAP/IMAPFilterRuleDeleteDialog'
import makeStyles from '@mui/styles/makeStyles'
import OtherActions from '../util/OtherActions'
import Spinner from '../loading/components/Spinner'
import { GenericError } from '../error-pages'
import { IMAPFilterRule, ServiceIMAPConfig } from '../../schema'
import { useIsWidthDown } from '../util/useWidth'
import { Add, Settings } from '@mui/icons-material'
import CompList from '../lists/CompList'
import { CompListItemText } from '../lists/CompListItems'

const IMAP_DESCRIPTION =
  'Monitor incoming emails and create alerts based on filter rules. Configure Gmail IMAP settings and define patterns to match specific emails.'

const query = gql`
  query ServiceIMAPQuery($serviceID: ID!) {
    service(id: $serviceID) {
      id
      name
      imapConfig {
        enabled
        host
        port
        username
        mailbox
        pollIntervalMinutes
        markAsRead
        deleteAfter
      }
      imapFilterRules {
        id
        name
        enabled
        fromPattern
        subjectPattern
        toPattern
        matchMode
        excludeReplies
      }
    }
  }
`

const useStyles = makeStyles(() => ({
  spacing: {
    marginBottom: 96,
  },
  configCard: {
    marginBottom: 16,
  },
}))

const sortRules = (a: IMAPFilterRule, b: IMAPFilterRule): number => {
  if (a.name.toLowerCase() < b.name.toLowerCase()) return -1
  if (a.name.toLowerCase() > b.name.toLowerCase()) return 1
  return 0
}

export default function ServiceIMAPPage(props: {
  serviceID: string
}): JSX.Element {
  const classes = useStyles()
  const isMobile = useIsWidthDown('md')
  const [showConfigDialog, setShowConfigDialog] = useState(false)
  const [showCreateRuleDialog, setShowCreateRuleDialog] = useState(false)
  const [editRuleID, setEditRuleID] = useState<string | null>(null)
  const [deleteRule, setDeleteRule] = useState<{
    id: string
    name: string
  } | null>(null)

  const [{ data, fetching, error }] = useQuery({
    query,
    variables: { serviceID: props.serviceID },
  })

  if (fetching && !data) return <Spinner />
  if (error) return <GenericError error={error.message} />

  const config: ServiceIMAPConfig | null = data.service.imapConfig
  const rules: IMAPFilterRule[] = data.service.imapFilterRules || []

  function renderConfig(): ReactElement {
    return (
      <Card className={classes.configCard}>
        <CardContent>
          <Grid container spacing={2} alignItems='center'>
            <Grid item xs>
              <Typography variant='h5' component='h2'>
                IMAP Configuration
              </Typography>
              {config ? (
                <React.Fragment>
                  <Typography color='textSecondary' gutterBottom>
                    {config.enabled ? (
                      <Chip label='Enabled' color='primary' size='small' />
                    ) : (
                      <Chip label='Disabled' size='small' />
                    )}
                  </Typography>
                  <Typography variant='body2'>
                    Host: {config.host}:{config.port}
                  </Typography>
                  <Typography variant='body2'>
                    Username: {config.username}
                  </Typography>
                  <Typography variant='body2'>
                    Mailbox: {config.mailbox}
                  </Typography>
                  <Typography variant='body2'>
                    Poll Interval: {config.pollIntervalMinutes} minutes
                  </Typography>
                </React.Fragment>
              ) : (
                <Typography color='textSecondary'>
                  IMAP email monitoring is not configured for this service.
                  Click &quot;Configure IMAP&quot; to get started.
                </Typography>
              )}
            </Grid>
            <Grid item>
              <Button
                variant='contained'
                onClick={() => setShowConfigDialog(true)}
                startIcon={<Settings />}
              >
                Configure IMAP
              </Button>
            </Grid>
          </Grid>
        </CardContent>
      </Card>
    )
  }

  function renderFilterRules(): ReactElement {
    const items = rules
      .slice()
      .sort(sortRules)
      .map((rule) => {
        const patterns: string[] = []
        if (rule.fromPattern) {
          patterns.push(`From: "${rule.fromPattern}" (${rule.matchMode})`)
        }
        if (rule.subjectPattern) {
          patterns.push(`Subject: "${rule.subjectPattern}" (${rule.matchMode})`)
        }
        if (rule.toPattern) {
          patterns.push(`To: "${rule.toPattern}" (${rule.matchMode})`)
        }

        return (
          <CompListItemText
            key={rule.id}
            title={
              <React.Fragment>
                {rule.name}{' '}
                {rule.enabled ? (
                  <Chip label='Enabled' color='primary' size='small' />
                ) : (
                  <Chip label='Disabled' size='small' />
                )}
              </React.Fragment>
            }
            subText={
              <React.Fragment>
                {patterns.map((p, i) => (
                  <React.Fragment key={i}>
                    {p}
                    <br />
                  </React.Fragment>
                ))}
                {rule.excludeReplies && (
                  <Typography variant='body2' color='textSecondary'>
                    Exclude reply emails
                  </Typography>
                )}
              </React.Fragment>
            }
            action={
              <OtherActions
                actions={[
                  {
                    label: 'Edit',
                    onClick: () => setEditRuleID(rule.id),
                  },
                  {
                    label: 'Delete',
                    onClick: () =>
                      setDeleteRule({ id: rule.id, name: rule.name }),
                  },
                ]}
              />
            }
          />
        )
      })

    return (
      <CompList
        data-cy='imap-filter-rules'
        emptyMessage='No filter rules exist for this service. Create a filter rule to start monitoring emails.'
        note='Filter rules define which emails should trigger alerts. At least one pattern (From, Subject, or To) must be specified.'
        hideActionOnMobile
        action={
          <Button
            variant='contained'
            onClick={() => setShowCreateRuleDialog(true)}
            startIcon={<Add />}
            disabled={!config || !config.enabled}
          >
            Create Filter Rule
          </Button>
        }
      >
        {items}
      </CompList>
    )
  }

  return (
    <React.Fragment>
      <Grid container spacing={2} className={classes.spacing}>
        <Grid item xs={12}>
          <Typography variant='body2' color='textSecondary' paragraph>
            {IMAP_DESCRIPTION}
          </Typography>
        </Grid>
        <Grid item xs={12}>
          {renderConfig()}
        </Grid>
        <Grid item xs={12}>
          <Card>
            <CardContent>{renderFilterRules()}</CardContent>
          </Card>
        </Grid>
      </Grid>
      {isMobile && config && config.enabled && (
        <CreateFAB
          onClick={() => setShowCreateRuleDialog(true)}
          title='Create Filter Rule'
        />
      )}
      <Suspense>
        {showConfigDialog && (
          <ServiceIMAPConfigDialog
            serviceID={props.serviceID}
            onClose={() => setShowConfigDialog(false)}
          />
        )}
        {showCreateRuleDialog && (
          <IMAPFilterRuleCreateDialog
            serviceID={props.serviceID}
            onClose={() => setShowCreateRuleDialog(false)}
          />
        )}
        {editRuleID && (
          <IMAPFilterRuleEditDialog
            filterRuleID={editRuleID}
            serviceID={props.serviceID}
            onClose={() => setEditRuleID(null)}
          />
        )}
        {deleteRule && (
          <IMAPFilterRuleDeleteDialog
            filterRuleID={deleteRule.id}
            filterRuleName={deleteRule.name}
            onClose={() => setDeleteRule(null)}
          />
        )}
      </Suspense>
    </React.Fragment>
  )
}
