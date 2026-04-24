import React from 'react'
import IconButton from '@mui/material/IconButton'
import MUIFavoriteIcon from '@mui/icons-material/Favorite'
import NotFavoriteIcon from '@mui/icons-material/FavoriteBorder'
import Tooltip from '@mui/material/Tooltip'
import _ from 'lodash'

interface SetFavoriteButtonProps {
  typeName: 'rotation' | 'service' | 'schedule' | 'escalationPolicy' | 'user'
  onClick: (event: React.FormEvent<HTMLFormElement>) => void
  isFavorite?: boolean
  loading?: boolean
}

export function FavoriteIcon(): JSX.Element {
  return (
    <MUIFavoriteIcon
      data-cy='fav-icon'
      sx={(theme) => ({
        color:
          theme.palette.mode === 'dark'
            ? theme.palette.primary.main
            : 'rgb(205, 24, 49)',
      })}
    />
  )
}

export function SetFavoriteButton({
  typeName,
  onClick,
  isFavorite,
  loading,
}: SetFavoriteButtonProps): JSX.Element | null {
  if (loading) {
    return null
  }

  const icon = isFavorite ? (
    <FavoriteIcon />
  ) : (
    <NotFavoriteIcon sx={{ color: 'inherit' }} />
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
        size='large'
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
