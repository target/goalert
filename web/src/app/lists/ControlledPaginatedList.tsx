import React, { ReactElement, useState } from 'react'
import {
  Checkbox,
  Grid,
  Icon,
  IconButton,
  makeStyles,
  Tooltip,
} from '@material-ui/core'
import {
  PaginatedList,
  PaginatedListItemProps,
  PaginatedListProps,
} from './PaginatedList'
import { useSelector } from 'react-redux'
import { urlKeySelector } from '../selectors/url'
import classnames from 'classnames'
import OtherActions from '../util/OtherActions'
import { ArrowDropDown } from '@material-ui/icons'
import Search from '../util/Search'

const useStyles = makeStyles({
  actionsContainer: {
    alignItems: 'center',
    display: 'flex',
    marginRight: 'auto',
    paddingLeft: '1em', // align with listItem icons
    width: 'fit-content',
  },
  checkbox: {
    marginTop: 4,
    marginBottom: 4,
  },
  controlsContainer: {
    alignItems: 'center',
    display: 'flex',
  },
  hover: {
    '&:hover': {
      cursor: 'pointer',
    },
  },
  popper: {
    opacity: 1,
  },
  search: {
    paddingLeft: '0.5em',
  },
})

export interface ControlledPaginatedListProps extends PaginatedListProps {
  checkboxActions?: ControlledPaginatedListAction[]
  filter?: ReactElement

  // if set, the search string param is ignored
  noSearch?: boolean

  // filters additional to search, set in the search text field
  searchAdornment?: ReactElement

  items: CheckboxItemsProps[] | PaginatedListItemProps[]
}

export interface ControlledPaginatedListAction {
  // icon for the action (e.g. X for close)
  icon: ReactElement

  // label to display (e.g. "Close alerts")
  label: string

  // Callback that will be passed a list of selected items
  onClick: (selectedIDs: (string | number)[]) => void
}

// used if checkBoxActions is set to true
export interface CheckboxItemsProps extends PaginatedListItemProps {
  // used to track checked items
  id: string | number

  /*
   * if false, checkbox will be disabled, and if already selected
   * it will be omitted from the action callback.
   *
   * defaults to true
   */
  selectable?: boolean
}

export default function ControlledPaginatedList(
  props: ControlledPaginatedListProps,
) {
  const classes = useStyles()
  const {
    checkboxActions,
    filter,
    noSearch,
    searchAdornment,
    items,
    ...listProps
  } = props

  const [_checkedItems, setCheckedItems] = useState<Array<string | number>>([])
  // covers the use case where an item may no longer be selectable after an update
  const checkedItems = _checkedItems.filter(id =>
    getSelectableIDs().includes(id),
  )
  const urlKey = useSelector(urlKeySelector)

  /*
   * ensures item type is of CheckboxItemsProps and not PaginatedListItemProps
   * checks all items against having no id present
   * returns true if all items have an id
   */
  function itemsHaveID(items: any): items is CheckboxItemsProps[] {
    return !items.some((i: CheckboxItemsProps) => !i.id)
  }

  function getSelectableIDs(): Array<string | number> {
    if (itemsHaveID(items)) {
      return items.filter(i => i.selectable !== false).map(i => i.id)
    }
    return []
  }

  function setAll() {
    setCheckedItems(getSelectableIDs())
  }

  function setNone() {
    setCheckedItems([])
  }

  function handleToggleSelectAll() {
    // if none are checked, set all
    if (checkedItems.length === 0) {
      setAll()
    } else {
      setNone()
    }
  }

  function getItems() {
    if (itemsHaveID(items)) {
      return items.map(item => ({ ...item, icon: getItemIcon(item) }))
    }

    return items
  }

  return (
    <React.Fragment>
      <Grid container item xs={12} justify='flex-end' alignItems='center'>
        {renderActions()}
        {filter}
        {!noSearch && (
          <Grid item className={classes.search}>
            <Search endAdornment={searchAdornment} />
          </Grid>
        )}
      </Grid>

      <Grid item xs={12}>
        <PaginatedList key={urlKey} {...listProps} items={getItems()} />
      </Grid>
    </React.Fragment>
  )

  function renderActions(): ReactElement | null {
    if (!checkboxActions) return null
    const itemIDs = getSelectableIDs()

    return (
      <Grid className={classes.actionsContainer} item container spacing={2}>
        <Grid item>
          <Checkbox
            className={classes.checkbox}
            checked={
              itemIDs.length === checkedItems.length && itemIDs.length > 0
            }
            data-cy='select-all'
            indeterminate={
              checkedItems.length > 0 && itemIDs.length !== checkedItems.length
            }
            onChange={handleToggleSelectAll}
          />
        </Grid>

        <Grid
          item
          className={classnames(classes.hover, classes.controlsContainer)}
          data-cy='checkboxes-menu'
        >
          <OtherActions
            icon={
              <Icon>
                <ArrowDropDown />
              </Icon>
            }
            actions={[
              {
                label: 'All',
                onClick: setAll,
              },
              {
                label: 'None',
                onClick: setNone,
              },
            ]}
            placement='right'
          />
        </Grid>

        {checkboxActions.map((a, idx) => (
          <Grid item key={idx}>
            <Tooltip
              title={a.label}
              placement='bottom'
              classes={{ popper: classes.popper }}
            >
              <IconButton onClick={() => a.onClick(checkedItems)}>
                {a.icon}
              </IconButton>
            </Tooltip>
          </Grid>
        ))}
      </Grid>
    )
  }

  function getItemIcon(item: CheckboxItemsProps) {
    if (!checkboxActions) return item.icon

    const checked = checkedItems.includes(item.id)

    return (
      <Checkbox
        checked={checked}
        data-cy={'item-' + item.id}
        disabled={item.selectable === false}
        onClick={e => {
          e.stopPropagation()
          e.preventDefault()

          if (checked) {
            setCheckedItems(checkedItems.filter(id => id !== item.id))
          } else {
            setCheckedItems([...checkedItems, item.id])
          }
        }}
      />
    )
  }
}
