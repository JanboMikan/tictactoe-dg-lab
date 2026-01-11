import { useState, useEffect } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Slider,
  Typography,
  Box,
  Alert,
} from '@mui/material';
import type { PlayerConfig } from '../../types/game';

interface SettingsDialogProps {
  open: boolean;
  onClose: () => void;
  onSave: (config: PlayerConfig) => void;
  currentConfig?: PlayerConfig;
}

const DEFAULT_CONFIG: PlayerConfig = {
  safe_min: 10,
  safe_max: 30,
  move_strength: 10,
  draw_strength: 15,
};

export const SettingsDialog = ({
  open,
  onClose,
  onSave,
  currentConfig = DEFAULT_CONFIG,
}: SettingsDialogProps) => {
  const [safeRange, setSafeRange] = useState<[number, number]>([
    currentConfig.safe_min,
    currentConfig.safe_max,
  ]);
  const [moveStrength, setMoveStrength] = useState(currentConfig.move_strength);
  const [drawStrength, setDrawStrength] = useState(currentConfig.draw_strength);
  const [error, setError] = useState<string | null>(null);

  // Update local state when currentConfig changes
  useEffect(() => {
    setSafeRange([currentConfig.safe_min, currentConfig.safe_max]);
    setMoveStrength(currentConfig.move_strength);
    setDrawStrength(currentConfig.draw_strength);
  }, [currentConfig]);

  const handleSafeRangeChange = (_event: Event, newValue: number | number[]) => {
    if (Array.isArray(newValue) && newValue.length === 2) {
      setSafeRange([newValue[0], newValue[1]]);

      // Auto-adjust move_strength and draw_strength to stay within range
      if (moveStrength < newValue[0]) {
        setMoveStrength(newValue[0]);
      } else if (moveStrength > newValue[1]) {
        setMoveStrength(newValue[1]);
      }

      if (drawStrength < newValue[0]) {
        setDrawStrength(newValue[0]);
      } else if (drawStrength > newValue[1]) {
        setDrawStrength(newValue[1]);
      }

      setError(null);
    }
  };

  const handleMoveStrengthChange = (_event: Event, newValue: number | number[]) => {
    if (typeof newValue === 'number') {
      setMoveStrength(newValue);
      setError(null);
    }
  };

  const handleDrawStrengthChange = (_event: Event, newValue: number | number[]) => {
    if (typeof newValue === 'number') {
      setDrawStrength(newValue);
      setError(null);
    }
  };

  const handleSave = () => {
    // Validate
    if (safeRange[0] >= safeRange[1]) {
      setError('Safe Min must be less than Safe Max');
      return;
    }

    if (moveStrength < safeRange[0] || moveStrength > safeRange[1]) {
      setError('Move Strength must be within Safe Range');
      return;
    }

    if (drawStrength < safeRange[0] || drawStrength > safeRange[1]) {
      setError('Draw Strength must be within Safe Range');
      return;
    }

    const config: PlayerConfig = {
      safe_min: safeRange[0],
      safe_max: safeRange[1],
      move_strength: moveStrength,
      draw_strength: drawStrength,
    };

    onSave(config);
    onClose();
  };

  const handleCancel = () => {
    // Reset to current config
    setSafeRange([currentConfig.safe_min, currentConfig.safe_max]);
    setMoveStrength(currentConfig.move_strength);
    setDrawStrength(currentConfig.draw_strength);
    setError(null);
    onClose();
  };

  return (
    <Dialog open={open} onClose={handleCancel} maxWidth="sm" fullWidth>
      <DialogTitle>Shock Intensity Settings</DialogTitle>
      <DialogContent>
        {error && (
          <Alert severity="error" sx={{ mb: 2 }}>
            {error}
          </Alert>
        )}

        <Box sx={{ mb: 4, mt: 2 }}>
          <Typography gutterBottom>
            Safe Range (Min - Max): {safeRange[0]} - {safeRange[1]}
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
            Set your comfortable intensity range (0-100). All shocks will be within this range.
          </Typography>
          <Slider
            value={safeRange}
            onChange={handleSafeRangeChange}
            valueLabelDisplay="auto"
            min={0}
            max={100}
            marks={[
              { value: 0, label: '0' },
              { value: 25, label: '25' },
              { value: 50, label: '50' },
              { value: 75, label: '75' },
              { value: 100, label: '100' },
            ]}
          />
        </Box>

        <Box sx={{ mb: 4 }}>
          <Typography gutterBottom>
            Move Shock Intensity: {moveStrength}
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
            Intensity when you make a move
          </Typography>
          <Slider
            value={moveStrength}
            onChange={handleMoveStrengthChange}
            valueLabelDisplay="auto"
            min={safeRange[0]}
            max={safeRange[1]}
            marks={[
              { value: safeRange[0], label: safeRange[0].toString() },
              { value: safeRange[1], label: safeRange[1].toString() },
            ]}
          />
        </Box>

        <Box sx={{ mb: 2 }}>
          <Typography gutterBottom>
            Draw Shock Intensity: {drawStrength}
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
            Intensity when the game ends in a draw
          </Typography>
          <Slider
            value={drawStrength}
            onChange={handleDrawStrengthChange}
            valueLabelDisplay="auto"
            min={safeRange[0]}
            max={safeRange[1]}
            marks={[
              { value: safeRange[0], label: safeRange[0].toString() },
              { value: safeRange[1], label: safeRange[1].toString() },
            ]}
          />
        </Box>
      </DialogContent>
      <DialogActions>
        <Button onClick={handleCancel}>Cancel</Button>
        <Button onClick={handleSave} variant="contained">
          Save
        </Button>
      </DialogActions>
    </Dialog>
  );
};
