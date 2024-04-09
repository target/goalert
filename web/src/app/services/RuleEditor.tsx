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
  CardActions,
  Box,
} from '@mui/material'
import { styles as globalStyles } from '../styles/materialStyles'
import makeStyles from '@mui/styles/makeStyles'
import RuleEditorConditionDialog from './RuleEditorConditionDialog'
import RuleEditorActionDialog from './RuleEditorActionDialog'
import MoreHorizIcon from '@mui/icons-material/MoreHoriz'
import { ActionInput, DestinationTypeInfo } from '../../schema'
import { useDynamicActionTypes } from '../util/RequireConfig'

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

const makeDefaultAction = (t: DestinationTypeInfo): ActionInput => ({
  dest: { type: t.type, values: [] },
  params: (t.dynamicParams || []).map((p) => ({
    paramID: p.paramID,
    expr: 'body.' + p.paramID,
  })),
})

export default function RuleEditor(): React.ReactNode {
  const classes = useStyles()
  const actTypes = useDynamicActionTypes()
  const [rules, setRules] = useState([
    {
      condition: 'foo == "bar" and baz < 3',
      actions: [makeDefaultAction(actTypes[0]), makeDefaultAction(actTypes[1])],
    },
  ])
  const [editCondition, setEditCondition] = useState<null | number>(null)
  const [editAction, setEditAction] = useState<null | {
    ruleIdx: number
    actionIdx: number
  }>(null)

  const actionLabel = (a: ActionInput): string =>
    actTypes.find((t) => t.type === a.dest.type)?.name || a.dest.type

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
          <Card key={idx} className={classes.margin} raised>
            <CardHeader
              component='h4'
              title={`Rule #${idx + 1}`}
              sx={{ margin: 0, paddingBottom: 0 }}
            />

            <CardContent>
              <Box
                sx={{
                  borderRadius: 1,
                  bgcolor: 'primary.dark',
                  padding: '16px',
                  marginBottom: '8px',
                }}
              >
                <Box display='flex' justifyContent='space-between'>
                  <Typography variant='h6' component='div'>
                    Condition
                  </Typography>
                  <Button
                    onClick={() =>
                      setEditCondition({ idx, value: r.condition })
                    }
                    variant='outlined'
                    color='primary'
                  >
                    Edit Condition
                  </Button>
                </Box>

                <Typography color='textSecondary'>
                  <Box
                    sx={{
                      borderRadius: 1,
                      padding: '0px 8px',
                      justifyContent: 'space-between',
                      display: 'flex',
                      alignItems: 'center',
                    }}
                  >
                    <Typography>{r.condition}</Typography>
                  </Box>
                </Typography>
              </Box>

              <Box
                sx={{
                  borderRadius: 1,
                  bgcolor: 'secondary.dark',
                  padding: '16px',
                }}
              >
                <Typography variant='h6' component='div'>
                  Actions{' '}
                  <Button
                    onClick={() => {
                      const newActionIndex = r.actions.length
                      setRules([
                        ...rules.slice(0, idx),
                        {
                          ...r,
                          actions: [
                            ...r.actions,
                            makeDefaultAction(actTypes[0]),
                          ],
                        },
                        ...rules.slice(idx + 1),
                      ])
                      setEditAction({ ruleIdx: idx, actionIdx: newActionIndex })
                    }}
                  >
                    Add Action
                  </Button>
                </Typography>
                {r.actions.length === 0 && (
                  <Typography color='textSecondary'>
                    <Box
                      sx={{
                        borderRadius: 1,
                        padding: '0px 8px',
                        justifyContent: 'space-between',
                        display: 'flex',
                        alignItems: 'center',
                      }}
                    >
                      <Typography>No Action/Drop Request</Typography>
                    </Box>
                  </Typography>
                )}
                {r.actions.map((a, i) => (
                  <Typography key={i} color='textSecondary'>
                    <Box
                      sx={{
                        borderRadius: 1,
                        padding: '0px 8px',
                        justifyContent: 'space-between',
                        display: 'flex',
                        alignItems: 'center',
                      }}
                    >
                      <Typography>{actionLabel(a)}</Typography>
                      <Button
                        onClick={() =>
                          setEditAction({ ruleIdx: idx, actionIdx: i })
                        }
                        endIcon={<MoreHorizIcon />}
                      />
                    </Box>
                  </Typography>
                ))}
              </Box>
            </CardContent>
            <CardActions sx={{ paddingTop: 0 }}>
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
            </CardActions>
          </Card>
        )
      })}

      <Card className={classes.margin} style={{ marginLeft: 0 }}>
        <CardHeader component='h3' title='Default' />
      </Card>
    </Grid>
  )
}
