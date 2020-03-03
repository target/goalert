import React, { ReactElement, useEffect } from 'react'
import {
  Checkbox,
  Grid,
  Icon,
  IconButton,
  makeStyles,
  Tooltip,
} from '@material-ui/core'
import Search from '../util/Search'
import { useDispatch, useSelector } from 'react-redux'
import { setCheckedItems as _setCheckedItems } from '../actions'
import { ArrowDropDown } from '@material-ui/icons'
import OtherActions from '../util/OtherActions'
import classnames from 'classnames'

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

export interface CheckboxActions {
  icon: ReactElement
  tooltip: string
  onClick: (items: string[]) => void
  ariaLabel?: string
  dataCy?: string
}

export default function ListControls(props: {
  // renders URL controlled search
  withSearch?: boolean
  searchAdornment?: ReactElement

  // list filter
  filter?: ReactElement

  // checkbox actions
  actions?: CheckboxActions[]

  itemIDs: Array<string>
}) {
  const { withSearch, searchAdornment, filter, actions, itemIDs } = props
  const classes = useStyles()

  const dispatch = useDispatch()
  const checkedItems = useSelector((state: any) => state.list.checkedItems)
  const setCheckedItems = (array: Array<any>) =>
    dispatch(_setCheckedItems(array))

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

  return (
    <Grid container item xs={12} justify='flex-end' alignItems='center'>
      {renderActions()}
      {filter}
      {withSearch && (
        <Grid item className={classes.search}>
          <Search endAdornment={searchAdornment} />
        </Grid>
      )}
    </Grid>
  )

  function renderActions(): ReactElement | null {
    if (!actions) return null

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

        {actions.map((a, idx) => (
          <Grid item key={idx}>
            <Tooltip
              title={a.tooltip}
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
