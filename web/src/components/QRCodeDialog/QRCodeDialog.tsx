import { Dialog, DialogTitle, DialogContent, Box, Typography, Button } from '@mui/material';
import { QRCodeSVG } from 'qrcode.react';

interface QRCodeDialogProps {
  open: boolean;
  onClose: () => void;
  dglabClientId: string;
  serverUrl: string;
}

export const QRCodeDialog = ({ open, onClose, dglabClientId, serverUrl }: QRCodeDialogProps) => {
  // Generate QR code content according to DG-LAB specification
  // Format: https://www.dungeon-lab.com/app-download.php#DGLAB-SOCKET#<socket_url>/<client_id>
  const qrContent = `https://www.dungeon-lab.com/app-download.php#DGLAB-SOCKET#${serverUrl}/${dglabClientId}`;

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>Connect DG-LAB Device</DialogTitle>
      <DialogContent>
        <Box
          sx={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            gap: 3,
            py: 2,
          }}
        >
          <Typography variant="body2" color="text.secondary" align="center">
            Scan this QR code with the DG-LAB app to connect your device
          </Typography>

          <Box
            sx={{
              p: 2,
              bgcolor: 'white',
              borderRadius: 2,
              display: 'inline-flex',
            }}
          >
            <QRCodeSVG value={qrContent} size={256} level="H" />
          </Box>

          <Box sx={{ width: '100%' }}>
            <Typography variant="caption" color="text.secondary" sx={{ wordBreak: 'break-all' }}>
              Client ID: {dglabClientId}
            </Typography>
          </Box>

          <Button variant="outlined" onClick={onClose} fullWidth>
            Close
          </Button>
        </Box>
      </DialogContent>
    </Dialog>
  );
};
