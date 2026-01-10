import { Container, Typography, Box } from '@mui/material'

function App() {
  return (
    <Container maxWidth="lg">
      <Box sx={{ my: 4 }}>
        <Typography variant="h3" component="h1" gutterBottom>
          DG-LAB Tic-Tac-Toe
        </Typography>
        <Typography variant="h6" color="text.secondary">
          郊狼井字棋游戏 - 前端已就绪
        </Typography>
      </Box>
    </Container>
  )
}

export default App
