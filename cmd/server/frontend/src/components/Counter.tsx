import { FC, useState } from 'react';

interface CounterProps {
	initialCount?: number;
	label: string;
}

const Counter: FC<CounterProps> = ({ initialCount = 0, label }) => {
	const [count, setCount] = useState < number > (initialCount);

	const increment = (): void => {
		setCount(prevCount => prevCount + 1);
	};

	const decrement = (): void => {
		setCount(prevCount => prevCount - 1);
	};

	return (
		<div className="bg-gray-800 p-6 rounded-lg shadow-lg w-64">
			<h3 className="text-xl font-semibold mb-2">{label}</h3>
			<p className="text-2xl font-bold mb-4">Count: {count}</p>
			<div className="flex justify-center space-x-4">
				<button
					onClick={decrement}
					className="bg-red-500 hover:bg-red-600 text-white font-bold py-2 px-4 rounded transition duration-200"
				>
					-
				</button>
				<button
					onClick={increment}
					className="bg-blue-500 hover:bg-blue-600 text-white font-bold py-2 px-4 rounded transition duration-200"
				>
					+
				</button>
			</div>
		</div>
	);
};

export default Counter;