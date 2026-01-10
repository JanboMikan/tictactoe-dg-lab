import { useEffect, useRef, useState, useCallback } from 'react';
import toast from 'react-hot-toast';
import type {
  GameMessage,
  RoomState,
  GameOverMessage,
  ShockEventMessage,
  ErrorMessage,
  PlayerConfig,
  JoinRoomMessage,
  UpdateDGLabIDMessage,
  UpdateConfigMessage,
  MoveMessage,
  PunishMessage,
} from '../types/game';

interface UseGameWebSocketProps {
  url: string;
  onRoomState?: (state: RoomState) => void;
  onGameOver?: (data: GameOverMessage) => void;
  onShockEvent?: (data: ShockEventMessage) => void;
  onError?: (error: string) => void;
}

export const useGameWebSocket = ({
  url,
  onRoomState,
  onGameOver,
  onShockEvent,
  onError,
}: UseGameWebSocketProps) => {
  const [isConnected, setIsConnected] = useState(false);
  const [roomState, setRoomState] = useState<RoomState | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<number | null>(null);
  const reconnectAttempts = useRef(0);
  const maxReconnectAttempts = 5;
  const shouldReconnect = useRef(true);
  const isReadyToSend = useRef(false);

  // Store callbacks in refs to avoid triggering reconnects
  const onRoomStateRef = useRef(onRoomState);
  const onGameOverRef = useRef(onGameOver);
  const onShockEventRef = useRef(onShockEvent);
  const onErrorRef = useRef(onError);

  // Update refs when callbacks change
  useEffect(() => {
    onRoomStateRef.current = onRoomState;
    onGameOverRef.current = onGameOver;
    onShockEventRef.current = onShockEvent;
    onErrorRef.current = onError;
  }, [onRoomState, onGameOver, onShockEvent, onError]);

  const connect = useCallback(() => {
    // Don't create multiple connections
    if (wsRef.current?.readyState === WebSocket.OPEN || wsRef.current?.readyState === WebSocket.CONNECTING) {
      console.log('WebSocket already connected or connecting');
      return;
    }

    console.log('Connecting to WebSocket:', url);
    const ws = new WebSocket(url);

    ws.onopen = () => {
      console.log('WebSocket connected');
      setIsConnected(true);
      reconnectAttempts.current = 0;
      // Wait a bit before allowing messages to be sent
      setTimeout(() => {
        isReadyToSend.current = true;
        console.log('WebSocket ready to send messages');
      }, 100);
      // Don't show toast on connect - too noisy
    };

    ws.onmessage = (event) => {
      try {
        const message: GameMessage = JSON.parse(event.data);
        console.log('Received message:', message);

        switch (message.type) {
          case 'room_state':
            setRoomState({
              board: message.board,
              turn: message.turn,
              players: message.players,
            });
            onRoomStateRef.current?.(message);
            break;

          case 'game_over':
            if (onGameOverRef.current) {
              onGameOverRef.current(message);
            }
            if (message.winner) {
              toast.success(`${message.winner} wins!`);
            } else {
              toast('Game ended in a draw', { icon: '🤝' });
            }
            break;

          case 'shock_event':
            if (onShockEventRef.current) {
              onShockEventRef.current(message);
            }
            const emoji =
              message.reason === 'punish' ? '⚡' : message.reason === 'move' ? '📍' : '💥';
            toast(`${emoji} ${message.target} received ${message.intensity} intensity shock`, {
              duration: 2000,
            });
            break;

          case 'error':
            const errorMsg = (message as ErrorMessage).error;
            console.error('Game error:', errorMsg);
            toast.error(errorMsg);
            onErrorRef.current?.(errorMsg);
            break;

          default:
            console.log('Unknown message type:', message);
        }
      } catch (error) {
        console.error('Failed to parse message:', error);
      }
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      // Don't show toast on error - will show on reconnect failure if needed
    };

    ws.onclose = () => {
      console.log('WebSocket disconnected');
      setIsConnected(false);
      isReadyToSend.current = false;
      wsRef.current = null;

      // Only auto-reconnect if we should (not manually disconnected)
      if (shouldReconnect.current && reconnectAttempts.current < maxReconnectAttempts) {
        const delay = Math.min(1000 * Math.pow(2, reconnectAttempts.current), 10000);
        console.log(`Reconnecting in ${delay}ms (attempt ${reconnectAttempts.current + 1})`);
        reconnectTimeoutRef.current = window.setTimeout(() => {
          reconnectAttempts.current++;
          connect();
        }, delay);
      } else if (reconnectAttempts.current >= maxReconnectAttempts) {
        toast.error('Failed to reconnect to server');
      }
    };

    wsRef.current = ws;
  }, [url]);

  const disconnect = useCallback(() => {
    shouldReconnect.current = false;
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
    setIsConnected(false);
  }, []);

  const sendMessage = useCallback(
    <T extends GameMessage>(message: T) => {
      if (wsRef.current?.readyState === WebSocket.OPEN && isReadyToSend.current) {
        const jsonMessage = JSON.stringify(message);
        console.log('Sending message:', jsonMessage);
        wsRef.current.send(jsonMessage);
      } else {
        console.error('WebSocket is not ready to send, message queued/dropped:', message.type);
        // Optionally: implement a message queue here
      }
    },
    []
  );

  // Convenience methods for common operations
  const joinRoom = useCallback(
    (playerName: string, roomId?: string) => {
      const message: JoinRoomMessage = {
        type: 'join_room',
        player_name: playerName,
        ...(roomId && { room_id: roomId }),
      };
      sendMessage(message);
    },
    [sendMessage]
  );

  const updateDGLabID = useCallback(
    (dglabClientId: string) => {
      const message: UpdateDGLabIDMessage = {
        type: 'update_dglab_id',
        dglab_client_id: dglabClientId,
      };
      sendMessage(message);
    },
    [sendMessage]
  );

  const updateConfig = useCallback(
    (config: PlayerConfig) => {
      const message: UpdateConfigMessage = {
        type: 'update_config',
        config,
      };
      sendMessage(message);
    },
    [sendMessage]
  );

  const makeMove = useCallback(
    (position: number) => {
      const message: MoveMessage = {
        type: 'move',
        position,
      };
      sendMessage(message);
    },
    [sendMessage]
  );

  const sendPunishment = useCallback(
    (percent: number, duration: number) => {
      const message: PunishMessage = {
        type: 'punish',
        percent,
        duration,
      };
      sendMessage(message);
    },
    [sendMessage]
  );

  useEffect(() => {
    shouldReconnect.current = true;
    connect();

    return () => {
      shouldReconnect.current = false;
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, [connect]);

  return {
    isConnected,
    roomState,
    joinRoom,
    updateDGLabID,
    updateConfig,
    makeMove,
    sendPunishment,
    sendMessage,
    disconnect,
  };
};
