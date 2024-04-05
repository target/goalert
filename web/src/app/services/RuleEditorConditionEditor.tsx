import React, { useState } from 'react'
import { Button, Box, Paper, Grid, Typography, Divider } from '@mui/material'
import ConditionRow from './RuleEditorConditionRow'

export interface Value {
  type: 'boolean' | 'number' | 'string' | 'object'
  data: boolean | number | string | object
}

let n: number = 1

const uniqueID = (): string => {
  return 'id_' + n++
}

export interface Condition {
  id: string
  key: string
  operator: string
  value: Value
}

const ConditionsEditor: React.FC = () => {
  const [conditions, setConditions] = useState<Condition[]>([
    {
      id: uniqueID(),
      key: '',
      operator: '',
      value: { type: 'string', data: '' },
    },
  ])

  const handleConditionUpdate = (
    conditionID: string,
    updatedCondition: Condition,
  ): void => {
    const newConditions = conditions.map((c) =>
      c.id === conditionID ? updatedCondition : c,
    )
    setConditions(newConditions)
  }

  const handleAddCondition = (): void => {
    setConditions([
      ...conditions,
      {
        id: uniqueID(),
        key: '',
        operator: '',
        value: { type: 'string', data: '' },
      },
    ])
  }

  const handleDeleteCondition = (conditionID: string): void => {
    const newConditions = conditions.filter((c) => c.id !== conditionID)
    setConditions(newConditions)
  }

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

  return (
    <Box
      display='flex'
      flexDirection='column'
      alignItems='center'
      justifyContent='center'
      mt='100px'
    >
      <Paper
        elevation={3}
        style={{ padding: '20px', width: '100%', maxWidth: '800px' }}
      >
        <Typography
          variant='h5'
          component='h2'
          style={{ textAlign: 'left', marginBottom: '20px' }}
        >
          Create/Edit Conditions
        </Typography>
        {conditions.map((condition, index) => (
          <React.Fragment key={index}>
            {index > 0 && <ConditionDivider />}
            <ConditionRow
              key={condition.id}
              initialCondition={condition}
              onConditionUpdate={(updatedCondition) =>
                handleConditionUpdate(condition.id, updatedCondition)
              }
              onDelete={() => handleDeleteCondition(condition.id)}
            />
          </React.Fragment>
        ))}
        <Box mt={2}>
          <Button
            variant='contained'
            color='primary'
            onClick={handleAddCondition}
          >
            Add Condition
          </Button>
        </Box>
      </Paper>
    </Box>
  )
}

export default ConditionsEditor
