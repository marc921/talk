import { FC } from 'react';
import { Link, useLocation } from 'react-router-dom';

interface NavbarProps {
  routes: { path: string; label: string }[];
}

const Navbar: FC<NavbarProps> = ({ routes }) => {
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
            {/* Dynamically generate buttons for each route */}
            {routes.map((route) => (
              <Link
                key={route.path}
                to={route.path}
                className={`px-3 py-2 rounded text-white ${isActive(route.path)}`}
              >
                {route.label}
              </Link>
            ))}
          </div>
        </div>
      </div>
    </nav>
  );
};

export default Navbar;