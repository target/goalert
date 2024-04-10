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
  IconButton,
} from '@mui/material'
import { styles as globalStyles } from '../styles/materialStyles'
import makeStyles from '@mui/styles/makeStyles'
import RuleEditorConditionDialog from './RuleEditorConditionDialog'

import { useDynamicActionTypes } from '../util/RequireConfig'
import RuleEditorActionsManager, {
  makeDefaultAction,
} from './RuleEditorActionsManager'
import { ActionInput } from '../../schema'

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
  const actTypes = useDynamicActionTypes()
  const [rules, setRules] = useState([
    {
      condition: 'foo == "bar" and baz < 3',
      actions: [makeDefaultAction(actTypes[0]), makeDefaultAction(actTypes[1])],
    },
  ])
  const [editCondition, setEditCondition] = useState<null | number>(null)
  const [defaultActions, setDefaultActions] = useState<ActionInput[]>([
    makeDefaultAction(actTypes[0]),
  ])

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
      </Suspense>

      {rules.map((r, idx) => {
        return (
          <Card
            key={idx}
            className={classes.margin}
            raised
            sx={{ marginRight: '16px' }}
          >
            <CardHeader
              component='h4'
              title={`Rule #${idx + 1}`}
              sx={{ margin: 0, paddingBottom: 0 }}
            />

            <CardContent>
              <Box
                sx={{
                  borderRadius: 1,
                  padding: '16px',
                  marginBottom: '16px',
                  outline: 'solid',
                  outlineWidth: '1px',
                  outlineColor: 'divider',
                }}
              >
                <Box
                  display='flex'
                  justifyContent='space-between'
                  marginBottom='8px'
                >
                  <Typography variant='h6' component='div'>
                    Condition
                  </Typography>
                  <Button
                    onClick={() =>
                      setEditCondition({ idx, value: r.condition })
                    }
                    variant='contained'
                    color='primary'
                    size='small'
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

              <RuleEditorActionsManager
                value={r.actions}
                onChange={(newActions) => {
                  setRules([
                    ...rules.slice(0, idx),
                    { ...r, actions: newActions },
                    ...rules.slice(idx + 1),
                  ])
                }}
                border
              />
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

      <Card
        className={classes.margin}
        style={{ marginLeft: 0, padding: '16px' }}
      >
        <RuleEditorActionsManager
          default
          value={defaultActions}
          onChange={setDefaultActions}
          border={false}
        />
      </Card>
    </Grid>
  )
}
