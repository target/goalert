import React from 'react'
import p from 'prop-types'
import List from '@material-ui/core/List'
import ListItem from '@material-ui/core/ListItem'
import ListItemText from '@material-ui/core/ListItemText'
import { FormContainer } from '../forms'
import { defaultTo } from 'lodash-es'
import { IntegerInput } from './AdminFieldComponents'
import withStyles from '@material-ui/core/styles/withStyles'

const styles = {
  listItem: {
    // leaves some room around fields without descriptions
    // 71px is the height of the checkbox field without w/o a desc
    minHeight: '71px',
    padding: '1em',
  },
  listItemText: {
    maxWidth: '50%',
  },
  listItemAction: {
    width: '50%',
    display: 'flex',
    justifyContent: 'flex-end',
  },
}

@withStyles(styles)
export default class AdminLimitsSection extends React.PureComponent {
  static propTypes = {
    fields: p.arrayOf(
      p.shape({
        id: p.string.isRequired,
        description: p.string,
        value: p.number.isRequired,
        label: p.string.isRequired,
      }),
    ),
  }

  static defaultProps = {
    fields: [],
  }

  render() {
    const { classes, fields, value } = this.props

    return (
      <FormContainer>
        <List>
          {fields.map((f, idx) => {
            return (
              <ListItem
                key={f.id}
                className={classes.listItem}
                divider={idx !== fields.length - 1}
              >
                <ListItemText
                  className={classes.listItemText}
                  primary={f.label}
                  secondary={f.description}
                />
                <div className={classes.listItemAction}>
                  <IntegerInput
                    name={f.id}
                    value={defaultTo(value[f.id], f.value)}
                    onChange={val =>
                      this.props.onChange(f.id, val === f.value ? null : val)
                    }
                  />
                </div>
              </ListItem>
            )
          })}
        </List>
      </FormContainer>
    )
  }
}
