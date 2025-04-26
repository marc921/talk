import { FC } from 'react';
import Counter from './Counter';

const Home: FC = () => {
  return (
    <div className="flex flex-col items-center justify-center bg-gray-900 text-white p-6">
      <main className="max-w-4xl mx-auto text-center">
        <h1 className="text-4xl font-bold mb-4">Hello, World!</h1>
        <p className="text-xl mb-8">
          This is my place on the Internet, to host some simple tools,
          <br />
          projects and have a bit of fun coding.
        </p>
      </main>
    </div>
  );
};

export default Home;