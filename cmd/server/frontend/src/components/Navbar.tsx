import { FC } from 'react';
import { Link, useLocation } from 'react-router-dom';

const Navbar: FC = () => {
  const location = useLocation();

  // Function to determine if a link is active
  const isActive = (path: string): string => {
    return location.pathname === path
      ? 'bg-blue-700'
      : 'hover:bg-gray-700';
  };

  return (
    <nav className="bg-gray-800 p-4">
      <div className="max-w-6xl mx-auto">
        <div className="flex justify-between items-center">
          <div className="text-white font-bold text-xl">
            React + Go App
          </div>

          <div className="flex space-x-4">
            <Link
              to="/"
              className={`px-3 py-2 rounded text-white ${isActive('/')}`}
            >
              Home
            </Link>
            <Link
              to="/qrcode"
              className={`px-3 py-2 rounded text-white ${isActive('/qrcode')}`}
            >
              QRCode Generator
            </Link>
            <Link
              to="/about"
              className={`px-3 py-2 rounded text-white ${isActive('/about')}`}
            >
              About Me
            </Link>
          </div>
        </div>
      </div>
    </nav>
  );
};

export default Navbar;