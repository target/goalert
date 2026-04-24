import React, { ReactElement, ReactNode, useState } from 'react'
import {
  Button,
  Card,
  Checkbox,
  Grid,
  IconButton,
  Tooltip,
} from '@mui/material'
import { Add, ArrowDropDown } from '@mui/icons-material'
import {
  PaginatedList,
  PaginatedListItemProps,
  PaginatedListProps,
} from './PaginatedList'
import { ListHeaderProps } from './ListHeader'
import OtherActions from '../util/OtherActions'
import Search from '../util/Search'
import { useURLKey } from '../actions'
import { useIsWidthDown } from '../util/useWidth'

export interface ControlledPaginatedListProps
  extends PaginatedListProps,
    ListHeaderProps {
  listHeader?: ReactNode

  checkboxActions?: ControlledPaginatedListAction[]
  secondaryActions?: ReactElement

  // if set, the search string param is ignored
  noSearch?: boolean

  // filters additional to search, set in the search text field
  searchAdornment?: ReactElement

  items: CheckboxItemsProps[] | PaginatedListItemProps[]

  renderCreateDialog?: (onClose: () => void) => JSX.Element | undefined

  createLabel?: string
  hideCreate?: boolean
  onSelectionChange?: (selectedIDs: (string | number)[]) => void
}

export interface ControlledPaginatedListAction {
  // icon for the action (e.g. X for close)
  icon: ReactElement

  // label to display (e.g. "Close")
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
): React.JSX.Element {
  const {
    checkboxActions,
    createLabel,
    secondaryActions,
    noSearch,
    renderCreateDialog,
    searchAdornment,
    items,
    listHeader,
    hideCreate,
    ...listProps
  } = props

  const [showCreate, setShowCreate] = useState(false)
  const isMobile = useIsWidthDown('md')

  /*
   * ensures item type is of CheckboxItemsProps and not PaginatedListItemProps
   * checks all items against having no id present
   * returns true if all items have an id
   */
  function itemsHaveID(
    items: CheckboxItemsProps[] | PaginatedListItemProps[],
  ): items is CheckboxItemsProps[] {
    return !items.some(
      (i: CheckboxItemsProps | PaginatedListItemProps) => !('id' in i),
    )
  }

  function getSelectableIDs(): Array<string | number> {
    if (itemsHaveID(items)) {
      return items.filter((i) => i.selectable !== false).map((i) => i.id)
    }
    return []
  }

  const [_checkedItems, _setCheckedItems] = useState<Array<string | number>>([])
  const setCheckedItems = (ids: Array<string | number>): void => {
    _setCheckedItems(ids)
    if (props.onSelectionChange) props.onSelectionChange(ids)
  }
  // covers the use case where an item may no longer be selectable after an update
  const checkedItems = _checkedItems.filter((id) =>
    getSelectableIDs().includes(id),
  )
  const urlKey = useURLKey()

  function setAll(): void {
    setCheckedItems(getSelectableIDs())
  }

  function setNone(): void {
    setCheckedItems([])
  }

  function handleToggleSelectAll(): void {
    // if none are checked, set all
    if (checkedItems.length === 0) {
      setAll()
    } else {
      setNone()
    }
  }

  function renderActions(): ReactElement | null {
    if (!checkboxActions) return null
    const itemIDs = getSelectableIDs()

    return (
      <Grid
        aria-label='List Checkbox Controls'
        container
        sx={{
          alignItems: 'center',
          display: 'flex',
          width: 'fit-content',
        }}
      >
        <Grid>
          <Checkbox
            sx={{ mt: '4px', mb: '4px', ml: '1em' }}
            checked={
              itemIDs.length === checkedItems.length && itemIDs.length > 0
            }
            data-cy='select-all'
            indeterminate={
              checkedItems.length > 0 && itemIDs.length !== checkedItems.length
            }
            onChange={handleToggleSelectAll}
            disabled={items.length === 0}
          />
        </Grid>

        <Grid
          sx={{
            alignItems: 'center',
            display: 'flex',
            '&:hover': { cursor: 'pointer' },
          }}
          data-cy='checkboxes-menu'
        >
          <OtherActions
            IconComponent={ArrowDropDown}
            disabled={items.length === 0}
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

        {checkedItems.length > 0 &&
          checkboxActions.map((a, idx) => (
            <Grid key={idx}>
              <Tooltip
                title={a.label}
                placement='bottom'
                slotProps={{ popper: { sx: { opacity: 1 } } }}
              >
                <IconButton
                  onClick={() => {
                    a.onClick(checkedItems)
                    setNone()
                  }}
                  size='large'
                >
                  {a.icon}
                </IconButton>
              </Tooltip>
            </Grid>
          ))}
      </Grid>
    )
  }

  function getItemIcon(
    item: CheckboxItemsProps,
  ): React.JSX.Element | undefined {
    if (!checkboxActions) return item.icon

    const checked = checkedItems.includes(item.id)

    return (
      <Checkbox
        checked={checked}
        data-cy={'item-' + item.id}
        disabled={item.selectable === false}
        onClick={(e) => {
          e.stopPropagation()
          e.preventDefault()

          if (checked) {
            setCheckedItems(checkedItems.filter((id) => id !== item.id))
          } else {
            setCheckedItems([...checkedItems, item.id])
          }
        }}
      />
    )
  }

  function getItems(): CheckboxItemsProps[] | PaginatedListItemProps[] {
    if (itemsHaveID(items)) {
      return items.map((item) => ({ ...item, icon: getItemIcon(item) }))
    }

    return items
  }

  return (
    <React.Fragment>
      <Grid
        size={12}
        container
        spacing={2}
        justifyContent='flex-start'
        alignItems='center'
      >
        {renderActions()}
        {!noSearch && (
          <Grid>
            <Search endAdornment={searchAdornment} />
          </Grid>
        )}
        {secondaryActions && <Grid>{secondaryActions}</Grid>}

        {!hideCreate && renderCreateDialog && !isMobile && (
          <Grid sx={{ ml: 'auto' }}>
            <Button
              variant='contained'
              startIcon={<Add />}
              onClick={() => setShowCreate(true)}
            >
              Create {createLabel}
            </Button>
            {showCreate && renderCreateDialog(() => setShowCreate(false))}
          </Grid>
        )}
      </Grid>

      <Grid size={12}>
        <Card>
          {listHeader}
          <PaginatedList key={urlKey} {...listProps} items={getItems()} />
        </Card>
      </Grid>
    </React.Fragment>
  )
}
