import React, { FC } from 'react';
import Counter from './Counter';

const Home: FC = () => {
  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-gray-900 text-white p-6">
      <main className="max-w-4xl mx-auto text-center">
        <h1 className="text-4xl font-bold mb-4">Hello World from React with TypeScript!</h1>
        <p className="text-xl mb-8">
          This TypeScript React app with Tailwind CSS is being served from a Go server.
        </p>
        
        <div className="flex flex-wrap justify-center gap-6 mt-8">
          <Counter label="Counter 1" initialCount={5} />
          <Counter label="Counter 2" initialCount={10} />
        </div>
      </main>
    </div>
  );
};

export default Home;