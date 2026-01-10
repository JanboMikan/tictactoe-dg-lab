// Game type definitions

export type PlayerSymbol = 'X' | 'O';
export type CellValue = PlayerSymbol | null;
export type Board = CellValue[];

export interface PlayerConfig {
  safe_min: number;
  safe_max: number;
  move_strength: number;
  draw_strength: number;
}

export interface PlayerInfo {
  connected: boolean;
  device_active: boolean;
}

export interface RoomState {
  board: Board;
  turn: string;
  players: Record<string, PlayerInfo>;
  game_over?: boolean;
  winner?: string | null;
}

// WebSocket Message Types
export type MessageType =
  | 'join_room'
  | 'update_dglab_id'
  | 'update_config'
  | 'move'
  | 'punish'
  | 'room_state'
  | 'game_over'
  | 'shock_event'
  | 'error';

export interface BaseMessage {
  type: MessageType;
}

export interface JoinRoomMessage extends BaseMessage {
  type: 'join_room';
  room_id?: string;
  player_name: string;
}

export interface UpdateDGLabIDMessage extends BaseMessage {
  type: 'update_dglab_id';
  dglab_client_id: string;
}

export interface UpdateConfigMessage extends BaseMessage {
  type: 'update_config';
  config: PlayerConfig;
}

export interface MoveMessage extends BaseMessage {
  type: 'move';
  position: number;
}

export interface PunishMessage extends BaseMessage {
  type: 'punish';
  percent: number;
  duration: number;
}

export interface RoomStateMessage extends BaseMessage {
  type: 'room_state';
  board: Board;
  turn: string;
  players: Record<string, PlayerInfo>;
  room_id?: string;
}

export interface GameOverMessage extends BaseMessage {
  type: 'game_over';
  winner: string | null;
  line?: number[];
}

export interface ShockEventMessage extends BaseMessage {
  type: 'shock_event';
  target: string;
  intensity: number;
  reason: 'move' | 'draw' | 'loss' | 'punish';
}

export interface ErrorMessage extends BaseMessage {
  type: 'error';
  error: string;
}

export type GameMessage =
  | JoinRoomMessage
  | UpdateDGLabIDMessage
  | UpdateConfigMessage
  | MoveMessage
  | PunishMessage
  | RoomStateMessage
  | GameOverMessage
  | ShockEventMessage
  | ErrorMessage;
