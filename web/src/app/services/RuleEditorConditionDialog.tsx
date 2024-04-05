import React, { useState } from 'react'

import FormDialog from '../dialogs/FormDialog'
import ConditionsEditor from './RuleEditorConditionEditor'
import { gql, useQuery } from 'urql'

const query = gql`
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

export default function RuleEditorConditionDialog(props: {
  expr: string
  onClose: (expr: string | null) => void
}): JSX.Element {
  const [value, setValue] = useState<string>(props.expr)
  const [parse] = useQuery({ query, variables: { expr: props.expr } })

  if (parse.error) throw new Error('too complex')
  const [cond, setCond] = useState(parse.data.expr.exprToCondition)

  return (
    <FormDialog
      maxWidth='sm'
      title='Edit Condition'
      onClose={() => props.onClose(null)}
      onSubmit={() => props.onClose(value)}
      form={<ConditionsEditor value={cond} onChange={setCond} />}
    />
  )
}
