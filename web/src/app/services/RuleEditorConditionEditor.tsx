import React from 'react'
import { Button, Box, Grid, Typography, Divider } from '@mui/material'
import ConditionRow from './RuleEditorConditionRow'
import { ConditionInput } from '../../schema'
import { InputFieldError } from '../util/errtypes'

const ConditionDivider: React.FC = () => {
  return (
    <Grid container alignItems='center' style={{ margin: '10px 0' }}>
      <Grid item xs>
        <Divider />
      </Grid>
      <Grid item px={2}>
        <Typography variant='subtitle1' color='primary'>
          And
        </Typography>
      </Grid>
      <Grid item xs>
        <Divider />
      </Grid>
    </Grid>
  )
}

export type ConditionEditorProps = {
  value: ConditionInput
  onChange: (value: ConditionInput) => void

  errors?: InputFieldError[]
}

function ConditionEditor(props: ConditionEditorProps): React.ReactNode {
  return (
    <React.Fragment>
      {props.value.clauses.map((clause, index) => (
        <React.Fragment key={index}>
          {index > 0 && <ConditionDivider />}
          <ConditionRow
            value={clause}
            onChange={(value) => {
              props.onChange({
                ...props.value,
                clauses: props.value.clauses.map((c, i) =>
                  i === index ? value : c,
                ),
              })
            }}
            onDelete={() => {
              props.onChange({
                ...props.value,
                clauses: props.value.clauses.filter((_, i) => i !== index),
              })
            }}
          />
        </React.Fragment>
      ))}
      <Box mt={2}>
        <Button
          variant='contained'
          color='secondary'
          onClick={() => {
            props.onChange({
              ...props.value,
              clauses: props.value.clauses.concat({
                field: '',
                operator: '==',
                value: '""',
                negate: false,
              }),
            })
          }}
        >
          Add Clause
        </Button>
      </Box>
    </React.Fragment>
  )
}

export default ConditionEditor
