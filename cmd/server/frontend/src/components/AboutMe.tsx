import React, { FC } from 'react';

const AboutMe: FC = () => {
  return (
    <div className="min-h-screen bg-gray-900 text-white p-6">
      <div className="max-w-3xl mx-auto">
        <h1 className="text-4xl font-bold mb-6">About Me</h1>
        
        <div className="bg-gray-800 rounded-lg p-6 shadow-lg">
          <p className="mb-4 text-lg">
            Hello! I'm Marc Brun, a software developer with a passion for building web applications. 
            I specialize in Go for backend development and React with TypeScript for frontend.
          </p>
          
          <p className="mb-4 text-lg">
            This example demonstrates how to serve a React application from a Go server, 
            combining the best of both worlds: Go's performance and React's interactive UI capabilities.
          </p>
          
          <h2 className="text-2xl font-semibold mt-6 mb-4">My Skills</h2>
          <ul className="list-disc pl-6 space-y-2">
            <li>Go (Golang) for backend development</li>
            <li>React with TypeScript for frontend</li>
            <li>Tailwind CSS for styling</li>
            <li>RESTful API design</li>
            <li>Docker and Kubernetes</li>
          </ul>
          
          <h2 className="text-2xl font-semibold mt-6 mb-4">Hobbies</h2>
          <p className="text-lg">
            When I'm not coding, I enjoy hiking, reading science fiction, and experimenting with new 
            programming languages and frameworks.
          </p>
        </div>
      </div>
    </div>
  );
};

export default AboutMe;