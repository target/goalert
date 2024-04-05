import React, { useEffect, useState } from 'react'
import {
  TextField,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  Grid,
  SelectChangeEvent,
  Divider,
  IconButton,
} from '@mui/material'
import DeleteIcon from '@mui/icons-material/Delete'
import { Condition, Value } from './RuleEditorConditionEditor'

interface ConditionRowProps {
  initialCondition: Condition
  onConditionUpdate: (updatedCondition: Condition) => void
  onDelete: () => void
}

interface KeyInputProps {
  value: string
  onChange: (event: React.ChangeEvent<HTMLInputElement>) => void
  onBlur: () => void
}

interface OperatorSelectProps {
  value: string
  onChange: (event: SelectChangeEvent<string>) => void
  onBlur: () => void
}

interface ValueInputProps {
  value: string
  onChange: (event: React.ChangeEvent<HTMLInputElement>) => void
  onBlur: () => void
  onTypeChange: (type: 'boolean' | 'number' | 'string' | 'object') => void
}

interface DeleteButtonProps {
  onClick: () => void
}

const KeyInput: React.FC<KeyInputProps> = ({ value, onChange, onBlur }) => (
  <Grid item xs>
    <TextField
      label='Key'
      variant='outlined'
      value={value}
      onChange={onChange}
      onBlur={onBlur}
      fullWidth
    />
  </Grid>
)

const OperatorSelect: React.FC<OperatorSelectProps> = ({
  value,
  onChange,
  onBlur,
}) => (
  <Grid item xs={2}>
    <FormControl variant='outlined' fullWidth>
      <InputLabel>Operator</InputLabel>
      <Select
        label='Operator'
        value={value}
        onChange={onChange}
        onBlur={onBlur}
      >
        <MenuItem value='=='>==</MenuItem>
        <MenuItem value='!='>!=</MenuItem>
        <Divider sx={{ my: 0.5 }} />
        <MenuItem value='<'>&lt;</MenuItem>
        <MenuItem value='<='>&lt;=</MenuItem>
        <MenuItem value='>'>&gt;</MenuItem>
        <MenuItem value='>='>&gt;=</MenuItem>
        <Divider sx={{ my: 0.5 }} />
        <MenuItem value='contains'>contains</MenuItem>
        <MenuItem value='not contains'>not contains</MenuItem>
      </Select>
    </FormControl>
  </Grid>
)

const ValueInput: React.FC<ValueInputProps> = ({
  value,
  onChange,
  onBlur,
  onTypeChange,
}) => {
  const [typeOptions, setTypeOptions] = useState<string[]>(['string'])

  useEffect(() => {
    const determineTypes = (input: string): string[] => {
      const types = ['string']
      if (input === '') return types
      if (!isNaN(Number(input))) types.unshift('number')
      if (input.toLowerCase() === 'true' || input.toLowerCase() === 'false')
        types.unshift('boolean')

      return types
    }

    setTypeOptions(determineTypes(value))
  }, [value])

  return (
    <React.Fragment>
      <Grid item xs>
        <TextField
          label='Value'
          variant='outlined'
          value={value}
          onChange={onChange}
          onBlur={onBlur}
          fullWidth
        />
      </Grid>
      <Grid item xs={1} marginLeft={-10}>
        <FormControl
          variant='outlined'
          size='small'
          sx={{ width: 120, marginLeft: -8 }}
        >
          <Select
            value={typeOptions[0]}
            onChange={(e) =>
              onTypeChange(
                e.target.value as 'boolean' | 'number' | 'string' | 'object',
              )
            }
          >
            {typeOptions.map((type) => (
              <MenuItem key={type} value={type}>
                {type}
              </MenuItem>
            ))}
          </Select>
        </FormControl>
      </Grid>
    </React.Fragment>
  )
}

const DeleteButton: React.FC<DeleteButtonProps> = ({ onClick }) => (
  <Grid item xs={1} ml={-2}>
    <IconButton onClick={onClick}>
      <DeleteIcon />
    </IconButton>
  </Grid>
)

const ConditionRow: React.FC<ConditionRowProps> = ({
  initialCondition,
  onConditionUpdate,
  onDelete,
}) => {
  const [localCondition, setLocalCondition] = useState(initialCondition)

  const handleLocalChange = (field: keyof Condition, value: string | Value) => {
    const updatedCondition = { ...localCondition, [field]: value }
    setLocalCondition(updatedCondition)
  }

  const handleBlur = (): void => {
    onConditionUpdate(localCondition)
  }

  const handleTypeChange = (
    type: 'boolean' | 'number' | 'string' | 'object',
  ): void => {
    const updatedCondition: Condition = {
      ...localCondition,
      value: { ...localCondition.value, type },
    }
    setLocalCondition(updatedCondition)
    onConditionUpdate(updatedCondition)
  }

  return (
    <Grid container spacing={2} alignItems='center'>
      <KeyInput
        value={localCondition.key}
        onChange={(e) => handleLocalChange('key', e.target.value)}
        onBlur={handleBlur}
      />
      <OperatorSelect
        value={localCondition.operator}
        onChange={(e) =>
          handleLocalChange('operator', e.target.value as string)
        }
        onBlur={handleBlur}
      />
      <ValueInput
        value={localCondition.value.data as string}
        onChange={(e) =>
          handleLocalChange('value', {
            ...localCondition.value,
            data: e.target.value,
          })
        }
        onBlur={handleBlur}
        onTypeChange={handleTypeChange}
        valueType={localCondition.value.type}
      />
      <DeleteButton onClick={onDelete} />
    </Grid>
  )
}

export default ConditionRow
