import React, { useState, useEffect } from 'react'
import List from '@mui/material/List'
import ListItem from '@mui/material/ListItem'
import ListItemText from '@mui/material/ListItemText'
import Typography from '@mui/material/Typography'
import { FormContainer } from '../forms'
import _ from 'lodash'
import { Theme } from '@mui/material/styles'
import makeStyles from '@mui/styles/makeStyles'
import {
  StringInput,
  StringListInput,
  IntegerInput,
  BoolInput,
} from './AdminFieldComponents'
import { ConfigValue } from '../../schema'
import { DEBOUNCE_DELAY } from '../config'

const components = {
  string: StringInput,
  stringList: StringListInput,
  integer: IntegerInput,
  boolean: BoolInput,
}

interface FieldProps extends ConfigValue {
  label: string
}

type Value = { [id: string]: string }
type OnChange = (id: string, value: null | string) => void

interface AdminSectionProps {
  headerNote?: string
  value: Value
  fields: FieldProps[]
  onChange: OnChange
}

interface AdminFieldProps extends FieldProps {
  index: number
  onChange: OnChange
  fieldsLength: number
}

const useStyles = makeStyles((theme: Theme) => ({
  listItem: {
    // leaves some room around fields without descriptions
    // 71px is the height of the checkbox field without w/o a desc
    minHeight: '71px',
    padding: '1em',
  },
  listItemText: {
    maxWidth: '50%',
    [theme.breakpoints.up('md')]: {
      maxWidth: '65%',
    },
  },
  listItemAction: {
    width: '50%',
    [theme.breakpoints.up('md')]: {
      width: '35%',
    },
    display: 'flex',
    justifyContent: 'flex-end',
  },
}))

function AdminField(props: AdminFieldProps): JSX.Element {
  const classes = useStyles()
  const Field = components[props.type]
  const [fieldValue, setFieldValue] = useState<string>(props.value)

  // debounce to set the value
  useEffect(() => {
    const t = setTimeout(() => {
      props.onChange(props.id, fieldValue)
    }, DEBOUNCE_DELAY)
    return () => clearTimeout(t)
  }, [fieldValue])

  return (
    <ListItem
      key={props.id}
      className={classes.listItem}
      divider={props.index !== props.fieldsLength - 1}
      selected={_.has(props.value, props.id)}
    >
      <ListItemText
        className={classes.listItemText}
        primary={props.label}
        secondary={props.description}
      />
      <div className={classes.listItemAction}>
        <Field
          name={props.id}
          value={fieldValue}
          password={props.password}
          onChange={(val) =>
            setFieldValue(val === props.value || val === null ? '' : val)
          }
        />
      </div>
    </ListItem>
  )
}

export default function AdminSection(props: AdminSectionProps): JSX.Element {
  // TODO: add 'reset to default' buttons
  const { fields, value, headerNote } = props

  return (
    <FormContainer>
      <List>
        {headerNote && (
          <ListItem>
            <ListItemText
              disableTypography
              secondary={
                <Typography color='textSecondary'>{headerNote}</Typography>
              }
              style={{ fontStyle: 'italic' }}
            />
          </ListItem>
        )}
        {fields.map((f: FieldProps, index: number) => (
          <AdminField
            key={f.id}
            {...f}
            index={index}
            value={value[f.id] ?? f.value}
            fieldsLength={fields.length}
            onChange={props.onChange}
          />
        ))}
      </List>
    </FormContainer>
  )
}
