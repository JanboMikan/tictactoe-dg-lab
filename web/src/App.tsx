import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { Layout } from './components/Layout/Layout';
import { HomePage } from './components/HomePage/HomePage';
import { GameRoom } from './components/GameRoom/GameRoom';

function App() {
  return (
    <BrowserRouter>
      <Layout>
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/room/:roomId" element={<GameRoom />} />
        </Routes>
      </Layout>
    </BrowserRouter>
  );
}

export default App;
