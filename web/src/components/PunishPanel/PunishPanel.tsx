import { useState } from 'react';
import {
  Paper,
  Typography,
  Slider,
  Button,
  Box,
  Stack,
  Alert,
} from '@mui/material';
import { Bolt as BoltIcon } from '@mui/icons-material';
import toast from 'react-hot-toast';

interface PunishPanelProps {
  minDuration: number; // From config (punishment_duration_min)
  maxDuration: number; // From config (punishment_duration_max)
  onPunish: (percent: number, duration: number) => void;
  disabled?: boolean;
}

export const PunishPanel = ({
  minDuration,
  maxDuration,
  onPunish,
  disabled = false,
}: PunishPanelProps) => {
  const [percent, setPercent] = useState(50);
  const [duration, setDuration] = useState((minDuration + maxDuration) / 2);

  const handlePercentChange = (_event: Event, newValue: number | number[]) => {
    if (typeof newValue === 'number') {
      setPercent(newValue);
    }
  };

  const handleDurationChange = (_event: Event, newValue: number | number[]) => {
    if (typeof newValue === 'number') {
      setDuration(newValue);
    }
  };

  const handleSendPunishment = () => {
    if (disabled) return;

    // Validate
    if (percent < 1 || percent > 100) {
      toast.error('Intensity must be between 1% and 100%');
      return;
    }

    if (duration < minDuration || duration > maxDuration) {
      toast.error(`Duration must be between ${minDuration}s and ${maxDuration}s`);
      return;
    }

    // Send punishment
    onPunish(percent, duration);

    // Show confirmation
    toast.success(`Punishment sent: ${percent}% intensity for ${duration.toFixed(1)}s! ⚡`, {
      duration: 3000,
      icon: '💥',
    });
  };

  return (
    <Paper
      elevation={3}
      sx={{
        p: 3,
        bgcolor: 'error.dark',
        color: 'error.contrastText',
        border: '2px solid',
        borderColor: 'error.main',
      }}
    >
      <Stack spacing={3}>
        <Box sx={{ textAlign: 'center' }}>
          <BoltIcon sx={{ fontSize: 48, mb: 1 }} />
          <Typography variant="h5" fontWeight="bold">
            Punishment Control
          </Typography>
          <Typography variant="body2" sx={{ mt: 1, opacity: 0.9 }}>
            You won! Send a shock to your opponent.
          </Typography>
        </Box>

        <Alert severity="warning" sx={{ bgcolor: 'warning.dark' }}>
          <Typography variant="body2">
            The actual intensity will be calculated based on your opponent's safe range.
          </Typography>
        </Alert>

        <Box>
          <Typography gutterBottom fontWeight="medium">
            Intensity: {percent}%
          </Typography>
          <Typography variant="body2" sx={{ mb: 2, opacity: 0.8 }}>
            Select punishment intensity (1% - 100%)
          </Typography>
          <Slider
            value={percent}
            onChange={handlePercentChange}
            valueLabelDisplay="auto"
            min={1}
            max={100}
            marks={[
              { value: 1, label: '1%' },
              { value: 25, label: '25%' },
              { value: 50, label: '50%' },
              { value: 75, label: '75%' },
              { value: 100, label: '100%' },
            ]}
            sx={{
              color: 'error.light',
              '& .MuiSlider-thumb': {
                bgcolor: 'error.contrastText',
              },
              '& .MuiSlider-track': {
                bgcolor: 'error.light',
              },
            }}
          />
        </Box>

        <Box>
          <Typography gutterBottom fontWeight="medium">
            Duration: {duration.toFixed(1)}s
          </Typography>
          <Typography variant="body2" sx={{ mb: 2, opacity: 0.8 }}>
            Select punishment duration ({minDuration}s - {maxDuration}s)
          </Typography>
          <Slider
            value={duration}
            onChange={handleDurationChange}
            valueLabelDisplay="auto"
            min={minDuration}
            max={maxDuration}
            step={0.5}
            marks={[
              { value: minDuration, label: `${minDuration}s` },
              { value: maxDuration, label: `${maxDuration}s` },
            ]}
            sx={{
              color: 'error.light',
              '& .MuiSlider-thumb': {
                bgcolor: 'error.contrastText',
              },
              '& .MuiSlider-track': {
                bgcolor: 'error.light',
              },
            }}
          />
        </Box>

        <Button
          variant="contained"
          size="large"
          startIcon={<BoltIcon />}
          onClick={handleSendPunishment}
          disabled={disabled}
          sx={{
            bgcolor: 'error.contrastText',
            color: 'error.dark',
            fontWeight: 'bold',
            '&:hover': {
              bgcolor: 'error.light',
              color: 'error.contrastText',
            },
            '&:disabled': {
              bgcolor: 'grey.600',
              color: 'grey.400',
            },
          }}
        >
          SEND SHOCK
        </Button>
      </Stack>
    </Paper>
  );
};
