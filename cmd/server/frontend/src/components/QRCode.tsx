import React, { useEffect, useState } from 'react';

const QRCodeGenerator: React.FC = () => {
  const [content, setContent] = useState('');
  const [size, setSize] = useState(256);
  const [color, setColor] = useState('black');
  const [qrCodeUrl, setQrCodeUrl] = useState('');

  useEffect(() => {
    if (!content.trim()) return;

    const url = `/api/v1/qrcode?content=${encodeURIComponent(content)}&size=${size}&color=${color}`;
    setQrCodeUrl(url);
  }, [content, size, color]);

  const handleDownload = async () => {
    if (!qrCodeUrl) return;

    try {
      const response = await fetch(qrCodeUrl);
      const blob = await response.blob();
      const downloadUrl = URL.createObjectURL(blob);

      const link = document.createElement('a');
      link.href = downloadUrl;
      link.download = 'qrcode.png';
      link.click();

      // Clean up the object URL
      URL.revokeObjectURL(downloadUrl);
    } catch (error) {
      console.error('Failed to download QR code:', error);
    }
  };

  return (
    <div className="min-h-screen max-w-3xl mx-auto p-6 flex flex-col items-center gap-4 ">
      <h1 className="text-4xl text-white font-bold mb-6">QR Code Generator</h1>
      <input
        type="text"
        id="content"
        value={content}
        onChange={(e) => setContent(e.target.value)}
        className="text-black px-3 py-2 border rounded-md"
        placeholder="Enter text or URL"
      />
      <div className="flex items-center gap-4 mt-4">
        <label htmlFor="size" className="text-white mt-4">
          Size (px):
        </label>
        <input
          type="number"
          id="size"
          value={size}
          onChange={(e) => setSize(Number(e.target.value))}
          className="text-black p-1 border rounded-md mt-2"
          min="100"
          max="1000"
        />
      </div>

      <div className="flex items-center gap-4 mt-4">
        <label htmlFor="color" className="text-white mt-4">
          Color:
        </label>
        <div className="flex gap-2 mt-2">
          {['black', 'red', 'green', 'blue', 'purple', 'orange'].map((colorOption) => (
            <button
              key={colorOption}
              onClick={() => setColor(colorOption)}
              className={`w-8 h-8 rounded-full ${color === colorOption ? 'ring-2 ring-white' : ''
                }`}
              style={{ backgroundColor: colorOption }}
              aria-label={`Select ${colorOption} color`}
            />
          ))}
        </div>
      </div>

      {qrCodeUrl && (
        <div className="mt-4 flex flex-col items-center">
          <img src={qrCodeUrl} alt="QR Code" className="border" />
          <button
            onClick={handleDownload}
            className="mt-2 text-blue-500 hover:underline"
          >
            Download QR Code
          </button>
        </div>
      )}
    </div>
  );
};

export default QRCodeGenerator;