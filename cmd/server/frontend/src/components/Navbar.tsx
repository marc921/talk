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
      <div className="max-w-6xl mx-auto flex justify-between items-center">
        {/* First route aligned on the left */}
        <div className="text-white font-bold text-xl">
          <Link
            to={routes[0].path}
            className={`px-3 py-2 rounded text-white hover:bg-gray-700`}
          >
            {routes[0].label}
          </Link>
        </div>

        {/* All other routes aligned on the right */}
        <div className="flex space-x-4">
          {routes.slice(1).map((route) => (
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
    </nav>
  );
};

export default Navbar;