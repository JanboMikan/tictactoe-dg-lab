import { Box, Paper } from '@mui/material';
import type { Board as BoardType, CellValue } from '../../types/game';

interface BoardProps {
  board: BoardType;
  onCellClick: (position: number) => void;
  disabled?: boolean;
  winningLine?: number[];
}

export const Board = ({ board, onCellClick, disabled, winningLine }: BoardProps) => {
  const renderCell = (value: CellValue, index: number) => {
    const isWinningCell = winningLine?.includes(index);

    return (
      <Paper
        key={index}
        elevation={2}
        sx={{
          aspectRatio: '1',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          cursor: disabled || value !== null ? 'default' : 'pointer',
          bgcolor: isWinningCell ? 'success.light' : 'background.paper',
          transition: 'all 0.2s',
          '&:hover': {
            bgcolor:
              !disabled && value === null
                ? 'action.hover'
                : isWinningCell
                ? 'success.light'
                : 'background.paper',
            transform: !disabled && value === null ? 'scale(1.05)' : 'none',
          },
        }}
        onClick={() => {
          if (!disabled && value === null) {
            onCellClick(index);
          }
        }}
      >
        <Box
          sx={{
            fontSize: { xs: '3rem', sm: '4rem', md: '5rem' },
            fontWeight: 'bold',
            color: value === 'X' ? 'primary.main' : value === 'O' ? 'secondary.main' : 'transparent',
            userSelect: 'none',
          }}
        >
          {value || ''}
        </Box>
      </Paper>
    );
  };

  return (
    <Box
      sx={{
        maxWidth: 600,
        width: '100%',
        mx: 'auto',
        my: 4,
      }}
    >
      <Box
        sx={{
          display: 'grid',
          gridTemplateColumns: 'repeat(3, 1fr)',
          gap: 2,
        }}
      >
        {board.map((cell, index) => renderCell(cell, index))}
      </Box>
    </Box>
  );
};
