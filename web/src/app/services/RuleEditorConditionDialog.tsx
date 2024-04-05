import React, { useEffect, useState } from 'react'

import FormDialog from '../dialogs/FormDialog'
import ConditionsEditor from './RuleEditorConditionEditor'
import { CombinedError, gql, useClient, useQuery } from 'urql'
import exp from 'constants'
import { FormControlLabel, Grid, Switch } from '@mui/material'

const parseQuery = gql`
  query ParseCondition($expr: String!) {
    expr {
      exprToCondition(input: { expr: $expr }) {
        clauses {
          field
          operator
          value
          negate
        }
      }
    }
  }
`

const exprQuery = gql`
  query CompileCondition($cond: ConditionInput!) {
    expr {
      conditionToExpr(input: { condition: $cond })
    }
  }
`

const noSuspense = { suspense: false }

export default function RuleEditorConditionDialog(props: {
  expr: string
  onClose: (expr: string | null) => void
}): JSX.Element {
  const [initialParse] = useQuery({
    query: parseQuery,
    variables: { expr: props.expr },
  })
  const [useAdvanced, setUseAdvanced] = useState<boolean>(!!initialParse.error)
  const [isTooComplex, setIsTooComplex] = useState<boolean>(
    !!initialParse.error,
  )
  const [error, setError] = useState<null | CombinedError>(null)

  const [cond, setCond] = useState(initialParse.data?.expr.exprToCondition)
  const [value, setValue] = useState<string>(props.expr)

  const [condToExpr] = useQuery({
    query: exprQuery,
    variables: { cond },
    pause: useAdvanced,
    context: noSuspense,
  })
  useEffect(() => {
    if (useAdvanced) return
    if (condToExpr.error) {
      setError(condToExpr.error)
      return
    }
    if (!condToExpr.data) return

    setValue(condToExpr.data.expr.conditionToExpr)
  }, [condToExpr.data?.expr.conditionToExpr])

  const [exprToCond] = useQuery({
    query: parseQuery,
    variables: { expr: value },
    pause: !useAdvanced,
    context: noSuspense,
  })
  useEffect(() => {
    if (exprToCond.error) {
      setIsTooComplex(true)
      return
    }

    setIsTooComplex(false)

    if (!useAdvanced) return
    if (!exprToCond.data) return
    setCond(exprToCond.data?.expr.exprToCondition)
  }, [exprToCond.data?.expr.exprToCondition])

  return (
    <FormDialog
      maxWidth='sm'
      errors={error ? [error] : undefined}
      title='Edit Condition'
      onClose={() => props.onClose(null)}
      onSubmit={() => props.onClose(value)}
      form={
        <Grid container>
          <FormControlLabel
            control={
              <Switch
                checked={isTooComplex || useAdvanced}
                onChange={(e) => setUseAdvanced(e.target.checked)}
                name='use-advanced'
                disabled={isTooComplex}
              />
            }
            label='Advanced'
          />

          {useAdvanced || isTooComplex ? (
            <textarea
              value={value}
              onChange={(e) => {
                setValue(e.target.value)
                setError(null)
              }}
            />
          ) : (
            <ConditionsEditor
              value={cond}
              onChange={(newCond) => {
                setCond(newCond)
                setError(null)
              }}
            />
          )}
        </Grid>
      }
    />
  )
}
