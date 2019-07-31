import { green, red, yellow } from '@material-ui/core/colors'

export default {
  noStatus: {
    borderLeft: '3px solid transparent',
  },
  statusOK: {
    borderLeft: '3px solid ' + green[600],
  },
  statusWarning: {
    borderLeft: '3px solid ' + yellow[600],
  },
  statusError: {
    borderLeft: '3px solid ' + red[600],
  },
}
