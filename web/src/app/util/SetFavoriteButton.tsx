import React from 'react'
import IconButton from '@material-ui/core/IconButton'
import MUIFavoriteIcon from '@material-ui/icons/Favorite'
import NotFavoriteIcon from '@material-ui/icons/FavoriteBorder'
import Tooltip from '@material-ui/core/Tooltip'
import { makeStyles } from '@material-ui/core/styles'
import _ from 'lodash'

interface SetFavoriteButtonProps {
  typeName: 'rotation' | 'service' | 'schedule' | 'escalationPolicy'
  onClick: (event: React.FormEvent<HTMLFormElement>) => void
  isFavorite?: boolean
  loading?: boolean
}

const useStyles = makeStyles({
  favorited: {
    color: 'rgb(205, 24, 49)',
  },
  notFavorited: {
    color: 'inherit',
  },
})

export function FavoriteIcon(): JSX.Element {
  const classes = useStyles()
  return <MUIFavoriteIcon data-cy='fav-icon' className={classes.favorited} />
}

export function SetFavoriteButton({
  typeName,
  onClick,
  isFavorite,
  loading,
}: SetFavoriteButtonProps): JSX.Element | null {
  const classes = useStyles()

  if (loading) {
    return null
  }

  const icon = isFavorite ? (
    <FavoriteIcon />
  ) : (
    <NotFavoriteIcon className={classes.notFavorited} />
  )

  const content = (
    <form
      onSubmit={(e) => {
        e.preventDefault()
        onClick(e)
      }}
    >
      <IconButton
        aria-label={
          isFavorite
            ? `Unset as a Favorite ${_.startCase(typeName).toLowerCase()}`
            : `Set as a Favorite ${_.startCase(typeName).toLowerCase()}`
        }
        type='submit'
        data-cy='set-fav'
      >
        {icon}
      </IconButton>
    </form>
  )

  return (
    <Tooltip title={isFavorite ? 'Unfavorite' : 'Favorite'} placement='top'>
      {content}
    </Tooltip>
  )
}
