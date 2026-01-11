import { useState, useEffect, useMemo, useRef } from 'react';
import { useParams, useLocation, useNavigate } from 'react-router-dom';
import {
  AppBar,
  Toolbar,
  Typography,
  Box,
  Paper,
  Button,
  Chip,
  Stack,
  IconButton,
  Alert,
} from '@mui/material';
import { Settings as SettingsIcon, Bluetooth as BluetoothIcon } from '@mui/icons-material';
import { v4 as uuidv4 } from 'uuid';
import toast from 'react-hot-toast';
import { useGameWebSocket } from '../../hooks/useGameWebSocket';
import { Board } from '../Board/Board';
import { QRCodeDialog } from '../QRCodeDialog/QRCodeDialog';
import { SettingsDialog } from '../SettingsDialog/SettingsDialog';
import { PunishPanel } from '../PunishPanel/PunishPanel';
import type { GameOverMessage, RoomState, PlayerConfig } from '../../types/game';

export const GameRoom = () => {
  const { roomId } = useParams<{ roomId: string }>();
  const location = useLocation();
  const navigate = useNavigate();
  const nickname = location.state?.nickname;

  const [qrDialogOpen, setQrDialogOpen] = useState(false);
  const [settingsDialogOpen, setSettingsDialogOpen] = useState(false);
  const [dglabClientId] = useState(() => uuidv4());
  const [gameOverData, setGameOverData] = useState<GameOverMessage | null>(null);
  const [localRoomState, setLocalRoomState] = useState<RoomState | null>(null);
  const [playerConfig, setPlayerConfig] = useState<PlayerConfig>({
    safe_min: 10,
    safe_max: 30,
    move_strength: 10,
    draw_strength: 15,
  });

  // Track if we've already joined the room
  const hasJoinedRoom = useRef(false);
  // Track previous device connection status to detect changes
  const previousDeviceActive = useRef<boolean>(false);

  // WebSocket configuration
  const wsUrl = useMemo(() => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.hostname;
    const port = import.meta.env.VITE_WS_PORT || '8080';
    return `${protocol}//${host}:${port}/ws/game`;
  }, []);

  const dglabServerUrl = useMemo(() => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.hostname;
    const port = import.meta.env.VITE_WS_PORT || '8080';
    return `${protocol}//${host}:${port}/ws/dglab`;
  }, []);

  const { isConnected, roomState, joinRoom, updateDGLabID, updateConfig, makeMove, sendPunishment } =
    useGameWebSocket({
      url: wsUrl,
      onRoomState: (state) => {
        console.log('Room state updated:', state);
        setLocalRoomState(state);
      },
      onGameOver: (data) => {
        setGameOverData(data);
      },
    });

  // Redirect if no nickname
  useEffect(() => {
    if (!nickname) {
      navigate('/');
    }
  }, [nickname, navigate]);

  // Join room on mount (only once)
  useEffect(() => {
    if (isConnected && roomId && nickname && !hasJoinedRoom.current) {
      console.log('Joining room:', roomId, 'as', nickname);
      hasJoinedRoom.current = true;

      // Wait for WebSocket to be fully ready before sending messages
      setTimeout(() => {
        joinRoom(nickname, roomId);

        // Update DG-LAB ID after joining
        setTimeout(() => {
          updateDGLabID(dglabClientId);
        }, 200);
      }, 200);
    }
  }, [isConnected, roomId, nickname]);

  // Reset hasJoinedRoom when disconnected
  useEffect(() => {
    if (!isConnected) {
      hasJoinedRoom.current = false;
    }
  }, [isConnected]);

  // Monitor device connection status and auto-close QR dialog when connected
  useEffect(() => {
    if (!nickname) return;

    const displayRoomState = roomState || localRoomState;
    const currentPlayer = displayRoomState?.players?.[nickname];
    const isDeviceActive = currentPlayer?.device_active || false;

    // Detect device connection: from false -> true
    if (!previousDeviceActive.current && isDeviceActive && qrDialogOpen) {
      // Device just connected, close QR dialog and show success toast
      setQrDialogOpen(false);
      toast.success('DG-LAB device connected successfully! 🎮', {
        duration: 3000,
        icon: '✅',
      });
    }

    // Detect device disconnection: from true -> false
    if (previousDeviceActive.current && !isDeviceActive) {
      // Device just disconnected, show warning toast
      toast.error('DG-LAB device disconnected! 📴', {
        duration: 4000,
        icon: '⚠️',
      });
    }

    // Update previous state
    previousDeviceActive.current = isDeviceActive;
  }, [roomState, localRoomState, nickname, qrDialogOpen]);

  const handleCellClick = (position: number) => {
    if (!gameOverData && roomState?.turn === nickname) {
      makeMove(position);
    }
  };

  const handleConnectDevice = () => {
    setQrDialogOpen(true);
  };

  const handleOpenSettings = () => {
    setSettingsDialogOpen(true);
  };

  const handleSaveSettings = (config: PlayerConfig) => {
    setPlayerConfig(config);
    updateConfig(config);
    toast.success('Settings saved successfully!');
  };

  const handlePunish = (percent: number, duration: number) => {
    sendPunishment(percent, duration);
  };

  const displayRoomState = roomState || localRoomState;
  const board = displayRoomState?.board || Array(9).fill(null);
  const currentTurn = displayRoomState?.turn;
  const players = displayRoomState?.players || {};

  const isMyTurn = currentTurn === nickname;
  const playerNames = Object.keys(players);
  const isWinner = gameOverData?.winner === nickname;
  const canPunish = isWinner && gameOverData !== null;

  return (
    <Box sx={{ flexGrow: 1 }}>
      <AppBar position="static">
        <Toolbar>
          <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
            Room: {roomId}
          </Typography>
          <IconButton color="inherit" edge="end" onClick={handleOpenSettings}>
            <SettingsIcon />
          </IconButton>
        </Toolbar>
      </AppBar>

      <Box sx={{ mt: 3 }}>
        {!isConnected && (
          <Alert severity="warning" sx={{ mb: 2 }}>
            Connecting to server...
          </Alert>
        )}

        {/* Player Info Area */}
        <Paper sx={{ p: 2, mb: 3 }}>
          <Stack direction="row" spacing={2} justifyContent="space-around" alignItems="center">
            {playerNames.length > 0 ? (
              playerNames.map((playerName) => {
                const playerInfo = players[playerName];
                const isCurrentPlayer = playerName === nickname;
                return (
                  <Box
                    key={playerName}
                    sx={{
                      textAlign: 'center',
                      p: 2,
                      borderRadius: 2,
                      bgcolor: isCurrentPlayer ? 'action.selected' : 'transparent',
                    }}
                  >
                    <Typography variant="h6" gutterBottom>
                      {playerName}
                      {isCurrentPlayer && ' (You)'}
                    </Typography>
                    <Stack direction="row" spacing={1} justifyContent="center">
                      <Chip
                        icon={<BluetoothIcon />}
                        label={playerInfo?.device_active ? 'Device Connected' : 'No Device'}
                        color={playerInfo?.device_active ? 'success' : 'default'}
                        size="small"
                      />
                      <Chip
                        label={playerInfo?.connected ? 'Online' : 'Offline'}
                        color={playerInfo?.connected ? 'primary' : 'default'}
                        size="small"
                      />
                    </Stack>
                  </Box>
                );
              })
            ) : (
              <Typography variant="body1" color="text.secondary">
                Waiting for players...
              </Typography>
            )}
          </Stack>
        </Paper>

        {/* Game Status */}
        <Box sx={{ textAlign: 'center', mb: 2 }}>
          {gameOverData ? (
            <Alert severity={gameOverData.winner === nickname ? 'success' : 'info'}>
              {gameOverData.winner
                ? `${gameOverData.winner} wins!`
                : "It's a draw!"}
            </Alert>
          ) : currentTurn ? (
            <Typography variant="h5" color={isMyTurn ? 'primary' : 'text.secondary'}>
              {isMyTurn ? "Your turn!" : `${currentTurn}'s turn`}
            </Typography>
          ) : (
            <Typography variant="h5" color="text.secondary">
              Waiting for game to start...
            </Typography>
          )}
        </Box>

        {/* Board */}
        <Board
          board={board}
          onCellClick={handleCellClick}
          disabled={!isMyTurn || !!gameOverData}
          winningLine={gameOverData?.line}
        />

        {/* Punishment Panel - Only show if winner and game is over */}
        {canPunish && (
          <Box sx={{ mt: 3, mb: 3 }}>
            <PunishPanel
              minDuration={1.0}
              maxDuration={10.0}
              onPunish={handlePunish}
              disabled={false}
            />
          </Box>
        )}

        {/* Action Area */}
        <Box sx={{ textAlign: 'center', mt: 3 }}>
          <Button variant="contained" onClick={handleConnectDevice} sx={{ mr: 2 }}>
            Connect Device
          </Button>
          {gameOverData && (
            <Button variant="outlined" onClick={() => navigate('/')}>
              Back to Home
            </Button>
          )}
        </Box>
      </Box>

      {/* QR Code Dialog */}
      <QRCodeDialog
        open={qrDialogOpen}
        onClose={() => setQrDialogOpen(false)}
        dglabClientId={dglabClientId}
        serverUrl={dglabServerUrl}
      />

      {/* Settings Dialog */}
      <SettingsDialog
        open={settingsDialogOpen}
        onClose={() => setSettingsDialogOpen(false)}
        onSave={handleSaveSettings}
        currentConfig={playerConfig}
      />
    </Box>
  );
};
