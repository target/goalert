import react, { ReactElement, useEffect, useState } from 'react'
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
import React from 'react'
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
  checkboxActions: ControlledPaginatedListAction[]
  filter?: ReactElement

  // if set, the search string param is ignored
  noSearch?: boolean

  // filters additional to search, set in the search text field
  searchAdornment?: ReactElement

  items: ControlledPaginatedListItemProps[]
}

export interface ControlledPaginatedListAction {
  // icon for the action (e.g. X for close)
  icon: ReactElement

  // label to display (e.g. "Close alerts")
  label: string

  // Callback that will be passed a list of selected items
  onClick: (selectedIDs: (string | number)[]) => void

  ariaLabel?: string
  dataCy?: string
}

export interface ControlledPaginatedListItemProps
  extends PaginatedListItemProps {
  // id to be passed to the action callback
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

  const [checkedItems, setCheckedItems] = useState<Array<string | number>>([])
  const urlKey = useSelector(urlKeySelector)
  const itemIDs = items.map(i => i.id)

  // reset checkedItems array on unmount
  useEffect(() => {
    return () => {
      setNone()
    }
  }, [])

  function setAll() {
    setCheckedItems(itemIDs)
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

  function getItems(): Array<ControlledPaginatedListItemProps> {
    return items.map(item => {
      const checked = checkedItems.includes(item.id)
      return {
        ...item,
        icon: (
          <Checkbox
            checked={checked}
            data-cy={'item-' + item.id}
            onClick={e => {
              e.stopPropagation()
              e.preventDefault()

              if (checked) {
                const idx = checkedItems.indexOf(item.id)
                const newItems = checkedItems.slice()
                newItems.splice(idx, 1)
                setCheckedItems(newItems)
              } else {
                setCheckedItems([...checkedItems, item.id])
              }
            }}
          />
        ),
      }
    })
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
        <PaginatedList
          key={urlKey}
          {...listProps}
          items={Boolean(checkboxActions) ? getItems() : items}
        />
      </Grid>
    </React.Fragment>
  )

  function renderActions(): ReactElement | null {
    if (!checkboxActions) return null

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
              <IconButton
                aria-label={a.ariaLabel}
                data-cy={a.dataCy}
                onClick={() => a.onClick(checkedItems)}
              >
                {a.icon}
              </IconButton>
            </Tooltip>
          </Grid>
        ))}
      </Grid>
    )
  }
}
