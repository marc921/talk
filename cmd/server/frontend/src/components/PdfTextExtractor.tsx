import React, { useState, useRef, ChangeEvent } from 'react';
import { API_URL } from '../App';

const PdfTextExtractor: React.FC = () => {
  const [fileName, setFileName] = useState<string>('');
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleFileChange = (e: ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files.length > 0) {
      const file = e.target.files[0];
      if (file.type !== 'application/pdf') {
        setError('Please select a PDF file');
        setFileName('');
        return;
      }
      setFileName(file.name);
      setError(null);
    } else {
      setFileName('');
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!fileInputRef.current?.files?.length) {
      setError('Please select a PDF file');
      return;
    }

    setIsLoading(true);
    setError(null);
    
    const formData = new FormData();
    formData.append('pdf', fileInputRef.current.files[0]);
    
    try {
      const response = await fetch(API_URL+'/api/v1/extract/pdf', {
        method: 'POST',
        body: formData,
      });
      
      if (!response.ok) {
        let errorMessage = 'Failed to extract text';
        try {
          const errorData = await response.json();
          errorMessage = errorData.error || errorMessage;
        } catch (e) {
          // If parsing JSON fails, use the default error message
        }
        throw new Error(errorMessage);
      }
      
      const text = await response.text();
      const blob = new Blob([text], { type: 'text/plain' });
      const downloadUrl = URL.createObjectURL(blob);
      
      const link = document.createElement('a');
      link.href = downloadUrl;
      link.download = `${fileName.replace(/\.[^/.]+$/, '')}.txt`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      
      URL.revokeObjectURL(downloadUrl);
      
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="max-w-md mx-auto p-6 bg-gray-900 rounded-lg shadow-md">
      <h1 className="text-2xl font-bold text-center mb-6 text-gray-100">PDF Text Extractor</h1>
      
      <form onSubmit={handleSubmit} className="space-y-6">
        <div className="flex flex-col space-y-2">
          <label className="text-sm font-medium text-gray-300">
            Select PDF File
          </label>
          <div className="flex items-center justify-center w-full">
            <label className="flex flex-col items-center justify-center w-full h-32 border-2 border-gray-600 border-dashed rounded-lg cursor-pointer bg-gray-800 hover:bg-gray-700">
              <div className="flex flex-col items-center justify-center pt-5 pb-6">
                <svg className="w-8 h-8 mb-4 text-gray-400" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 20 16">
                  <path stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M13 13h3a3 3 0 0 0 0-6h-.025A5.56 5.56 0 0 0 16 6.5 5.5 5.5 0 0 0 5.207 5.021C5.137 5.017 5.071 5 5 5a4 4 0 0 0 0 8h2.167M10 15V6m0 0L8 8m2-2 2 2"/>
                </svg>
                <p className="mb-2 text-sm text-gray-400">
                  <span className="font-semibold">Click to upload</span> or drag and drop
                </p>
                <p className="text-xs text-gray-500">PDF files only</p>
                {fileName && <p className="mt-2 text-sm text-gray-300 truncate max-w-xs">{fileName}</p>}
              </div>
              <input 
                ref={fileInputRef}
                id="dropzone-file" 
                type="file" 
                className="hidden" 
                name="pdf"
                accept="application/pdf"
                onChange={handleFileChange}
              />
            </label>
          </div>
        </div>
        
        {error && (
          <div className="p-3 bg-red-900 text-red-300 rounded-md text-sm">
            {error}
          </div>
        )}
        
        <button
          type="submit"
          className="w-full py-2 px-4 bg-blue-700 hover:bg-blue-800 text-white font-medium rounded-lg text-sm transition-colors disabled:bg-blue-500"
          disabled={isLoading || !fileName}
        >
          {isLoading ? (
            <div className="flex items-center justify-center">
              <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              Extracting Text...
            </div>
          ) : 'Extract Text'}
        </button>
      </form>
    </div>
  );
};

export default PdfTextExtractor;