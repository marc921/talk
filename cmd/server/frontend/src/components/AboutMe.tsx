import { FC } from 'react';
import engineerImage from '../assets/engineer.png';

const AboutMe: FC = () => {
  return (
    <div className="min-h-screen mx-auto bg-gray-900 text-white p-6">
      <h1 className="text-center text-4xl font-bold mb-6">About Me</h1>
      <div className="flex items-start mb-6">
        <img
          src={engineerImage}
          alt="Engineer"
          className="w-80 rounded-lg mr-6"
        />
        <div className="bg-gray-800 rounded-lg p-6 shadow-lg">

          <p className="mb-4 text-lg">
            Hello! I'm Marc Brun, a Senior Backend and DevOps Engineer with a passion for building scalable and robust systems.
            I specialize in Go for backend development and have extensive experience with Kubernetes and modern cloud infrastructure.
          </p>

          <p className="mb-4 text-lg">
            My motivation lies in creating efficient, secure, and scalable solutions that empower businesses to grow and innovate.
            I enjoy tackling complex challenges and continuously improving code quality using modern tools and frameworks.
          </p>

          <h2 className="text-2xl font-semibold mt-6 mb-4">My Skills</h2>
          <ul className="list-disc pl-6 space-y-2">
            <li>Go for backend development</li>
            <li>Docker & Kubernetes for container orchestration</li>
            <li>Cloud infrastructure (AWS, GCP, Scaleway, Terraform)</li>
            <li>React for frontend</li>
          </ul>

          <h2 className="text-2xl font-semibold mt-6 mb-4">Hobbies</h2>
          <p className="text-lg">
            Outside of work, I enjoy hiking, reading science fiction, and exploring new programming languages and technologies.
          </p>
          <div className="flex items-center justify-between mt-6">
            <a
              href="https://www.linkedin.com/in/marc-brun-175218112/"
              target="_blank"
              rel="noopener noreferrer"
              className="text-blue-400 hover:text-blue-600"
            >
              LinkedIn
            </a>
            <a
              href="https://github.com/marc921"
              target="_blank"
              rel="noopener noreferrer"
              className="text-blue-400 hover:text-blue-600"
            >
              GitHub
            </a>
          </div>
        </div>
      </div>
    </div>
  );
};

export default AboutMe;