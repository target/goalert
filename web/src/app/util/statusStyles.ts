import { green, red, yellow } from '@material-ui/core/colors'

export const colors = {
  noStatus: 'transparent',
  statusOK: green[600],
  statusWarning: yellow[600],
  statusError: red[600],
}

export default {
  noStatus: {
    borderLeft: '3px solid transparent',
  },
  statusOK: {
    borderLeft: '3px solid ' + colors.statusOK,
  },
  statusWarning: {
    borderLeft: '3px solid ' + colors.statusWarning,
  },
  statusError: {
    borderLeft: '3px solid ' + colors.statusError,
  },
}
