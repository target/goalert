export const colors = {
  noStatus: 'transparent',
  statusOK: '#00e676',
  statusWarning: '#ffd602',
  statusError: '#ff324d',
}

export default {
  noStatus: {
    borderLeft: '3px solid ' + colors.noStatus,
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
