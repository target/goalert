export const styles = theme => ({
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
  asLink: {
    color: 'blue',
    cursor: 'pointer',
    textDecoration: 'underline',
  },
  block: {
    display: 'inline-block',
  },
  cancelButton: {
    color: 'rgba(0,0,0, 0.54)',
  },
  defaultFlex: {
    flex: '0 1 auto',
  },
  dndDragging: {
    backgroundColor: '#ebebeb',
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
  highlightRow: {
    backgroundColor: theme.palette.primary['100'],
  },
  error: {
    color: theme.palette.error['700'],
  },
  selectedOption: {
    backgroundColor: 'rgba(0, 0, 0, 0.12)',
  },
  trashIcon: {
    color: '#666',
    cursor: 'pointer',
    float: 'right',
  },
  warningColor: {
    color: '#FFD602',
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
  // used for react-router Link components
  nav: {
    color: theme.palette.primary['500'],
    textDecoration: 'none',
    display: 'block',
    '&:hover': {
      textDecoration: 'none',
    },
  },
  navSelected: {
    backgroundColor: '#ebebeb',
    borderRight: '3px solid ' + theme.palette.primary['500'],
  },
  // parent container must have position: relative
  topRightActions: {
    position: 'absolute',
    top: '0.7em',
    right: '0.7em',
  },
  searchFieldBox: {
    borderRadius: 4,
    backgroundColor: theme.palette.common.white,
    border: '1px solid #ced4da',
    fontSize: 16,
    padding: '10px 12px',
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
