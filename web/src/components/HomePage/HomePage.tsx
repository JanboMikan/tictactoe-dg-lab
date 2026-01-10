import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Box,
  Paper,
  TextField,
  Button,
  Typography,
  Divider,
  Stack,
} from '@mui/material';
import { v4 as uuidv4 } from 'uuid';

export const HomePage = () => {
  const navigate = useNavigate();
  const [nickname, setNickname] = useState('');
  const [roomId, setRoomId] = useState('');

  const handleCreateRoom = () => {
    if (!nickname.trim()) {
      alert('Please enter a nickname');
      return;
    }
    const newRoomId = uuidv4().substring(0, 6).toUpperCase();
    navigate(`/room/${newRoomId}`, { state: { nickname: nickname.trim() } });
  };

  const handleJoinRoom = () => {
    if (!nickname.trim()) {
      alert('Please enter a nickname');
      return;
    }
    if (!roomId.trim()) {
      alert('Please enter a room ID');
      return;
    }
    navigate(`/room/${roomId.trim().toUpperCase()}`, { state: { nickname: nickname.trim() } });
  };

  const handleKeyPress = (e: React.KeyboardEvent, action: 'create' | 'join') => {
    if (e.key === 'Enter') {
      if (action === 'create') {
        handleCreateRoom();
      } else {
        handleJoinRoom();
      }
    }
  };

  return (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        minHeight: '80vh',
      }}
    >
      <Paper
        elevation={3}
        sx={{
          p: 4,
          maxWidth: 500,
          width: '100%',
        }}
      >
        <Typography variant="h4" component="h1" gutterBottom align="center">
          DG-LAB Tic-Tac-Toe
        </Typography>
        <Typography variant="body2" color="text.secondary" align="center" sx={{ mb: 4 }}>
          Shock-Tac-Toe: Where every move has consequences ⚡
        </Typography>

        <Stack spacing={3}>
          <TextField
            fullWidth
            label="Nickname"
            variant="outlined"
            value={nickname}
            onChange={(e) => setNickname(e.target.value)}
            onKeyPress={(e) => handleKeyPress(e, 'create')}
            placeholder="Enter your nickname"
            inputProps={{ maxLength: 20 }}
          />

          <Button
            fullWidth
            variant="contained"
            size="large"
            onClick={handleCreateRoom}
            disabled={!nickname.trim()}
          >
            Create New Room
          </Button>

          <Divider>
            <Typography variant="body2" color="text.secondary">
              OR
            </Typography>
          </Divider>

          <TextField
            fullWidth
            label="Room ID"
            variant="outlined"
            value={roomId}
            onChange={(e) => setRoomId(e.target.value.toUpperCase())}
            onKeyPress={(e) => handleKeyPress(e, 'join')}
            placeholder="Enter room ID"
            inputProps={{ maxLength: 6 }}
          />

          <Button
            fullWidth
            variant="outlined"
            size="large"
            onClick={handleJoinRoom}
            disabled={!nickname.trim() || !roomId.trim()}
          >
            Join Room
          </Button>
        </Stack>
      </Paper>
    </Box>
  );
};
