import React, { useState } from 'react';
import { API_URL } from '../App';

const HTMLToMarkdownConverter = () => {
  const [file, setFile] = useState<File | undefined>(undefined);
  const [markdown, setMarkdown] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const selectedFile = e.target.files?.[0];
    if (selectedFile) {
      setFile(selectedFile);
      setError('');
    }
  };

  const convertHtmlToMarkdown = async () => {
    setLoading(true);
    setError('');
    setMarkdown('');

    if (!file) {
      setError('Please select a file to convert');
      setLoading(false);
      return;
    }

    try {
      const formData = new FormData();
      formData.append('html', file);

      const response = await fetch(API_URL + '/api/v1/html-to-markdown', {
        method: 'POST',
        body: formData,
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Conversion failed');
      }

      const data = await response.json();
      setMarkdown(data.markdown);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="mx-10 p-6">
      <div className="max-w-md mx-auto p-6">
        <h1 className="text-2xl font-bold text-center mb-6 text-white">HTML to Markdown Converter</h1>

        <div className="space-y-6">
          <div className="flex flex-col space-y-2">
            <label className="text-sm font-medium text-gray-300">Select HTML File</label>
            <div className="flex items-center justify-center w-full">
              <label className="flex flex-col items-center justify-center w-full h-32 border-2 border-gray-600 border-dashed rounded-lg cursor-pointer bg-gray-800 hover:bg-gray-700">
                <div className="flex flex-col items-center justify-center pt-5 pb-6">
                  <svg
                    className="w-8 h-8 mb-4 text-gray-400"
                    aria-hidden="true"
                    xmlns="http://www.w3.org/2000/svg"
                    fill="none"
                    viewBox="0 0 20 16"
                  >
                    <path
                      stroke="currentColor"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth="2"
                      d="M13 13h3a3 3 0 0 0 0-6h-.025A5.56 5.56 0 0 0 16 6.5 5.5 5.5 0 0 0 5.207 5.021C5.137 5.017 5.071 5 5 5a4 4 0 0 0 0 8h2.167M10 15V6m0 0L8 8m2-2 2 2"
                    />
                  </svg>
                  <p className="mb-2 text-sm text-gray-400">
                    <span className="font-semibold">Click to upload</span> or drag and drop
                  </p>
                  <p className="text-xs text-gray-400">HTML files only</p>
                  {file && <p className="mt-2 text-sm text-gray-300">{file.name}</p>}
                </div>
                <input
                  type="file"
                  accept=".html,.htm"
                  onChange={handleFileChange}
                  className="hidden"
                />
              </label>
            </div>
          </div>

          {error && (
            <div className="p-3 bg-red-900 text-red-300 rounded-md text-sm">{error}</div>
          )}

          <button
            onClick={convertHtmlToMarkdown}
            disabled={loading || !file}
            className="w-full py-2 px-4 bg-blue-700 hover:bg-blue-800 text-white font-medium rounded-lg text-sm transition-colors disabled:bg-blue-500"
          >
            {loading ? 'Converting...' : 'Convert to Markdown'}
          </button>
        </div>
      </div>

      {markdown && (
        <div className="mt-8">
          <div className="flex items-center gap-8 mb-2">
            <h2 className="text-xl font-semibold text-gray-300">Converted Markdown</h2>
            <button
              onClick={() => {
                const blob = new Blob([markdown], { type: 'text/markdown' });
                const url = URL.createObjectURL(blob);
                const a = document.createElement('a');
                a.href = url;
                a.download = 'converted.md';
                document.body.appendChild(a);
                a.click();
                document.body.removeChild(a);
                URL.revokeObjectURL(url);
              }}
              className="py-1 px-3 bg-green-700 text-white text-sm rounded-lg hover:bg-green-800"
            >
              Download
            </button>
          </div>
          <div className="border border-gray-600 rounded-lg p-4 bg-gray-900">
            <pre className="whitespace-pre-wrap break-words text-sm text-gray-300">{markdown}</pre>
          </div>
        </div>
      )}
    </div>
  );
};

export default HTMLToMarkdownConverter;