import React from 'react'
import IconButton from '@material-ui/core/IconButton'
import FavoriteIcon from '@material-ui/icons/Favorite'
import NotFavoriteIcon from '@material-ui/icons/FavoriteBorder'
import Tooltip from '@material-ui/core/Tooltip'
import { makeStyles } from '@material-ui/core/styles'

interface SetFavoriteButtonProps {
  typeName: 'rotation' | 'service' | 'schedule'
  isFavorite?: boolean
  loading: boolean
  onClick: (event: React.FormEvent<HTMLFormElement>) => void
}

const useStyles = makeStyles({
  favorited: {
    color: 'rgb(205, 24, 49)',
  },
  notFavorited: {
    color: 'inherit',
  },
})

export function SetFavoriteButton({
  typeName,
  isFavorite,
  loading,
  onClick,
}: SetFavoriteButtonProps): JSX.Element | null {
  const classes = useStyles()

  if (loading) {
    return null
  }

  const icon = isFavorite ? <FavoriteIcon /> : <NotFavoriteIcon />

  const content = (
    <form
      onSubmit={(e) => {
        e.preventDefault()
        onClick(e)
      }}
    >
      <IconButton
        className={isFavorite ? classes.favorited : classes.notFavorited}
        aria-label={
          isFavorite
            ? `Unset as a Favorite ${typeName}`
            : `Set as a Favorite ${typeName}`
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
