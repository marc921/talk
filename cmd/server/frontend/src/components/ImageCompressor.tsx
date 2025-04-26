import React, { useState, useRef, ChangeEvent } from 'react';
import { API_URL } from '../App';

const ImageCompressor: React.FC = () => {
  const [quality, setQuality] = useState<number>(75);
  const [fileName, setFileName] = useState<string>('');
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleQualityChange = (e: ChangeEvent<HTMLInputElement>) => {
    setQuality(parseInt(e.target.value));
  };

  const handleFileChange = (e: ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files.length > 0) {
      setFileName(e.target.files[0].name);
      setError(null);
    } else {
      setFileName('');
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!fileInputRef.current?.files?.length) {
      setError('Please select an image');
      return;
    }

    setIsLoading(true);
    setError(null);
    
    const formData = new FormData();
    formData.append('image', fileInputRef.current.files[0]);
    formData.append('quality', quality.toString());
    
    try {
      const response = await fetch(API_URL+'/api/v1/compress/image', {
        method: 'POST',
        body: formData,
      });
      
      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to compress image');
      }
      
      // Create a download link for the compressed image
      const blob = await response.blob();
      const downloadUrl = URL.createObjectURL(blob);
      
      const link = document.createElement('a');
      link.href = downloadUrl;
      link.download = `compressed_${fileName}`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="max-w-md mx-auto p-6">
      <h1 className="text-2xl font-bold text-center mb-6 text-white">Image Compression Service</h1>
      
      <form onSubmit={handleSubmit} className="space-y-6">
        <div className="flex flex-col space-y-2">
          <label className="text-sm font-medium text-gray-300">
            Select Image (PNG or JPEG)
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
                <p className="text-xs text-gray-400">PNG or JPEG</p>
                {fileName && <p className="mt-2 text-sm text-gray-300">{fileName}</p>}
              </div>
              <input 
                ref={fileInputRef}
                id="dropzone-file" 
                type="file" 
                className="hidden" 
                name="image"
                accept="image/png, image/jpeg"
                onChange={handleFileChange}
              />
            </label>
          </div>
        </div>
        
        <div className="space-y-2">
          <div className="flex justify-between">
            <label htmlFor="quality" className="text-sm font-medium text-gray-300">
              Quality
            </label>
            <span className="text-sm text-gray-400">{quality}%</span>
          </div>
          <input
            type="range"
            id="quality"
            name="quality"
            min="1"
            max="100"
            value={quality}
            onChange={handleQualityChange}
            className="w-full h-2 bg-gray-600 rounded-lg appearance-none cursor-pointer"
          />
          <div className="flex justify-between text-xs text-gray-400">
            <span>Lower Size</span>
            <span>Higher Quality</span>
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
          {isLoading ? 'Compressing...' : 'Compress Image'}
        </button>
      </form>
    </div>
  );
};

export default ImageCompressor;