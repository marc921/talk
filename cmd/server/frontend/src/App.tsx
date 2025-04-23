import { FC } from 'react';
import { Route, BrowserRouter as Router, Routes } from 'react-router-dom';
import AboutMe from './components/AboutMe';
import Home from './components/Home';
import Navbar from './components/Navbar';
import QRCodeGenerator from './components/QRCode';

const App: FC = () => {
  return (
    <Router>
      <div className="flex flex-col min-h-screen bg-gray-900">
        <Navbar />
        <div className="flex-grow">
          <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/about" element={<AboutMe />} />
            <Route path="/qrcode" element={<QRCodeGenerator />} />
          </Routes>
        </div>
      </div>
    </Router>
  );
};

export default App;