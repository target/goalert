import React from 'react'
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
import { ClauseInput } from '../../schema'

interface DeleteButtonProps {
  onClick: () => void
}

interface OperatorSelectProps {
  value: string
  onChange: (event: SelectChangeEvent<string>) => void
}
const OperatorSelect: React.FC<OperatorSelectProps> = ({ value, onChange }) => (
  <Grid item xs={2}>
    <FormControl variant='outlined' fullWidth>
      <InputLabel>Operator</InputLabel>
      <Select label='Operator' value={value} onChange={onChange}>
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

const opTypes: Record<string, string[]> = {
  '==': ['string', 'number', 'boolean'],
  '!=': ['string', 'number', 'boolean'],
  '<': ['number'],
  '<=': ['number'],
  '>': ['number'],
  '>=': ['number'],
  contains: ['string'],
  'not contains': ['string'],
}

interface ValueInputProps {
  value: string
  typeName: string
  operator: string
  onChange: (event: React.ChangeEvent<HTMLInputElement>) => void
  onTypeChange: (type: 'boolean' | 'number' | 'string') => void
}

const ValueInput: React.FC<ValueInputProps> = ({
  value,
  typeName,
  operator,
  onChange,
  onTypeChange,
}) => {
  const typeOptions = opTypes[operator] || []

  return (
    <React.Fragment>
      <Grid item xs>
        <TextField
          label='Value'
          variant='outlined'
          value={value}
          onChange={onChange}
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
            value={typeName}
            onChange={(e) =>
              onTypeChange(e.target.value as 'boolean' | 'number' | 'string')
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

function stringToType(
  value: string,
  type: 'boolean' | 'number' | 'string',
): boolean | number | string {
  switch (type) {
    case 'boolean':
      return value === 'true'
    case 'number':
      return Number(value)
    case 'string':
      return value
  }
}

interface ConditionRowProps {
  value: ClauseInput
  onChange: (newValue: ClauseInput) => void
  onDelete: () => void
}

const ConditionRow: React.FC<ConditionRowProps> = (props) => {
  const value = JSON.parse(props.value.value)
  const typeName = typeof value as 'string' | 'number' | 'boolean'

  function handleValueChange(newValueString: string): void {
    props.onChange({
      ...props.value,
      value: JSON.stringify(stringToType(newValueString, typeName)),
    })
  }

  return (
    <Grid container spacing={2} alignItems='center'>
      <Grid item xs>
        <TextField
          label='Key'
          variant='outlined'
          value={props.value.field}
          onChange={(e) =>
            props.onChange({ ...props.value, field: e.target.value })
          }
          fullWidth
        />
      </Grid>

      <OperatorSelect
        value={props.value.operator}
        onChange={(e) =>
          props.onChange({
            ...props.value,
            operator: e.target.value,
            negate: e.target.value.startsWith('not '),
          })
        }
      />

      <ValueInput
        value={value.toString()}
        onChange={(e) => handleValueChange(e.target.value)}
        onTypeChange={(newType) =>
          props.onChange({
            ...props.value,
            value: JSON.stringify(stringToType(value.toString(), newType)),
          })
        }
        typeName={typeName}
        operator={props.value.operator}
      />
      <DeleteButton onClick={props.onDelete} />
    </Grid>
  )
}

export default ConditionRow
