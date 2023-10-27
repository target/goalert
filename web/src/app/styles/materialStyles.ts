import { Theme } from '@mui/material/styles'
import { StyleRules } from '@mui/styles'

export const styles = (theme: Theme): StyleRules => ({
  // used on details pages on desktop
  card: {
    width: '75%',
  },
  // used on details pages on mobile
  cardFull: {
    width: '100%',
  },
  // used on details pages that contain cards (grid items or divs)
  cardContainer: {
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
  },
  // Used to reduce padding between header and content
  cardHeader: {
    paddingBottom: 0,
    margin: 0,
  },
  block: {
    display: 'inline-block',
  },
  defaultFlex: {
    flex: '0 1 auto',
  },
  // align reset and done buttons to the right in grid item
  filterActions: {
    display: 'flex',
    justifyContent: 'flex-end',
  },
  // used to act as a container for the page to center children components
  fullWidthDiv: {
    display: 'flex',
    justifyContent: 'center',
    width: '100%',
  },
  hidden: {
    visibility: 'hidden',
  },
  overflowVisible: {
    overflow: 'visible',
  },
  tableCell: {
    whiteSpace: 'normal',
    wordWrap: 'break-word',
    padding: '0.5em',
  },
  paper: {
    padding: '0.5em',
  },
  tableCardContent: {
    padding: 0,
  },
  error: {
    color: theme.palette.error.main,
  },
  selectedOption: {
    backgroundColor: 'rgba(0, 0, 0, 0.12)',
  },
  dialogWidth: {
    minWidth: '33vw',
  },
  vertCenterAlign: {
    margin: 'auto',
    marginLeft: 0,
  },
  grow: {
    flexGrow: 1,
  },
  // used for Link components
  nav: {
    borderRadius: 4,
    margin: 8,
    textDecoration: 'none',
    display: 'block',
    '& p, span': {
      lineHeight: '1.375rem',
      color: theme.palette.text.secondary, // todo: use text.primary with baseline mui palette
    },
    '&:focus': {
      backgroundColor: 'rgba(0, 0, 0, 0.12)',
    },
    '&:hover': {
      borderRadius: 4,
      margin: 8,
      textDecoration: 'none',
      overflow: 'hidden',
    },
  },
  navSelected: {
    textDecoration: 'none',
    display: 'block',
    '& p, span': {
      lineHeight: '1.375rem',
      color: theme.palette.text.secondary, // todo: use text.primary with baseline mui palette
    },
    '&:focus': {
      backgroundColor: 'rgba(0, 0, 0, 0.12)',
    },
    '&:hover': {
      borderRadius: 4,
      margin: 8,
      textDecoration: 'none',
      overflow: 'hidden',
    },
    backgroundColor: theme.palette.primary.main + '1f', // 12% opacity
    borderRadius: 4,
    margin: 8,

    // text and icon
    '& p, svg': {
      color: theme.palette.primary.main,
    },
  },
  // parent container must have position: relative
  topRightActions: {
    position: 'absolute',
    top: '0.7em',
    right: '0.7em',
  },
  // use on grid items except the last one per page
  mobileGridSpacing: {
    marginBottom: '1em',
  },
  srOnly: {
    clip: 'rect(1px, 1px, 1px, 1px)',
    overflow: 'hidden',
    height: 1,
    width: 1,
  },
})
