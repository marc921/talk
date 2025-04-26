import { FC } from 'react';
import { Route, BrowserRouter as Router, Routes } from 'react-router-dom';
import AboutMe from './components/AboutMe';
import Home from './components/Home';
import Navbar from './components/Navbar';
import QRCodeGenerator from './components/QRCode';
import ImageCompressor from './components/ImageCompressor';
import PdfTextExtractor from './components/PdfTextExtractor';
import HTMLToMarkdownConverter from './components/HTMLToMarkdown';

export const API_URL = 'http://localhost:8080';

// Define routes in a configuration array
const routes = [
  { path: '/', element: <Home />, label: 'marcbrun.eu' },
  { path: '/about', element: <AboutMe />, label: 'About Me' },
  { path: '/qrcode', element: <QRCodeGenerator />, label: 'QR Code Generator' },
  { path: '/compress/image', element: <ImageCompressor />, label: 'Image Compressor' },
  { path: '/extract/pdf', element: <PdfTextExtractor />, label: 'PDF Text Extractor' },
  { path: '/convert/html', element: <HTMLToMarkdownConverter />, label: 'HTML to Markdown Converter' },
];

const App: FC = () => {
  return (
    <Router>
      <div className="flex flex-col min-h-screen bg-gray-900">
        {/* Pass routes as props to Navbar */}
        <Navbar routes={routes} />
        <div className="flex-grow">
          <Routes>
            {/* Dynamically generate routes */}
            {routes.map((route) => (
              <Route key={route.path} path={route.path} element={route.element} />
            ))}
          </Routes>
        </div>
      </div>
    </Router>
  );
};

export default App;