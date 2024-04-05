import React, { Suspense, useState } from 'react'
import {
  Grid,
  Card,
  CardHeader,
  Chip,
  Theme,
  Typography,
  CardContent,
  Button,
} from '@mui/material'
import { styles as globalStyles } from '../styles/materialStyles'
import makeStyles from '@mui/styles/makeStyles'
import RuleEditorConditionDialog from './RuleEditorConditionDialog'
import RuleEditorActionDialog from './RuleEditorActionDialog'

const useStyles = makeStyles((theme: Theme) => {
  const { cardHeader } = globalStyles(theme)

  return {
    margin: {
      marginTop: theme.spacing(4),
      marginLeft: theme.spacing(2),
    },
    padding: {
      padding: theme.spacing(2),
    },
    cardHeader,
  }
})

export default function RuleEditor(): React.ReactNode {
  const classes = useStyles()
  const [rules, setRules] = useState([
    {
      condition: 'foo == "bar" and baz < 3',
      actions: [{ type: 'createAlert' }, { type: 'slackMessage' }],
    },
  ])
  const [editCondition, setEditCondition] = useState<null | number>(null)
  const [editAction, setEditAction] = useState<null | {
    ruleIdx: number
    actionIdx: number
  }>(null)

  return (
    <Grid item xs={12}>
      <Card className={classes.padding}>
        <Grid item xs container direction='column'>
          <Grid item>
            <CardHeader
              title='My Custom Key'
              //   titleTypographyProps={{
              //     'data-cy': 'title',
              //     variant: 'h5',
              //     component: 'h1',
              //   }}
              //   subheaderTypographyProps={{
              //     'data-cy': 'subheader',
              //     variant: 'body1',
              //   }}
            />
          </Grid>

          <Grid item container spacing={1} sx={{ pl: '16px', pr: '16px' }}>
            <Grid item>
              <Chip label='service = Example Service' />
            </Grid>
          </Grid>
        </Grid>
      </Card>

      <Suspense>
        {editCondition && (
          <RuleEditorConditionDialog
            expr={rules[editCondition.idx].condition}
            onClose={(newCond) => {
              setEditCondition(null)
              if (newCond === null) return
              rules[editCondition.idx].condition = newCond
            }}
          />
        )}
        {editAction && editAction.actionIdx !== -1 && (
          <RuleEditorActionDialog
            onClose={(newAction) => {
              setEditAction(null)
              if (newAction === null) return
              // TODO: update action
            }}
          />
        )}
      </Suspense>

      {rules.map((r, idx) => {
        return (
          <Card key={idx} className={classes.margin}>
            <CardHeader component='h3' title={`Rule #${idx + 1}`} />
            <CardContent>
              <h3>Condition</h3>
              <Typography color='textSecondary'>
                {r.condition}
                <Button
                  onClick={() => setEditCondition({ idx, value: r.condition })}
                >
                  Edit Condition
                </Button>
              </Typography>
              <h3>Actions</h3>
              {r.actions.map((a, i) => (
                <Typography key={i} color='textSecondary'>
                  {a.type}
                  <Button
                    onClick={() =>
                      setEditAction({ ruleIdx: idx, actionIdx: i })
                    }
                  >
                    Edit Action
                  </Button>
                </Typography>
              ))}

              <Button
                onClick={() => {
                  // insert new rule after current rule
                  setRules([
                    ...rules.slice(0, idx + 1),
                    { condition: 'foo == "bar"', actions: [] },
                    ...rules.slice(idx + 1),
                  ])
                }}
              >
                Add Rule
              </Button>
              <Button
                onClick={() => {
                  setRules(rules.filter((rule, i) => i !== idx))
                }}
              >
                Delete Rule
              </Button>
            </CardContent>
          </Card>
        )
      })}

      <Card className={classes.margin} style={{ marginLeft: 0 }}>
        <CardHeader component='h3' title='Default' />
      </Card>
    </Grid>
  )
}
