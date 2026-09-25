import React, { useEffect, useRef, useState } from 'react';
import { Html5Qrcode } from 'html5-qrcode';
import { Camera, X, RefreshCw, Upload } from 'lucide-react';

export default function QRScanner({ onScan, onClose }) {
  const [error, setError] = useState(null);
  const [isScanning, setIsScanning] = useState(false);
  const scannerRef = useRef(null);
  const fileInputRef = useRef(null);

  useEffect(() => {
    const scannerId = 'qr-reader-container';
    const html5QrCode = new Html5Qrcode(scannerId);
    scannerRef.current = html5QrCode;

    const config = {
      fps: 10,
      qrbox: { width: 250, height: 250 },
      aspectRatio: 1.0,
    };

    html5QrCode
      .start(
        { facingMode: 'environment' },
        config,
        (decodedText) => {
          handleSuccess(decodedText);
        },
        (errorMessage) => {
          // Continuous scanning ignore routine frame misses
        }
      )
      .then(() => {
        setIsScanning(true);
      })
      .catch((err) => {
        console.warn('Camera start error:', err);
        setError('Camera permission denied or camera not available. You can also upload a QR image.');
      });

    return () => {
      if (scannerRef.current && scannerRef.current.isScanning) {
        scannerRef.current
          .stop()
          .then(() => scannerRef.current.clear())
          .catch((e) => console.error('Stop error', e));
      }
    };
  }, []);

  const handleSuccess = (rawText) => {
    // If QR contains full URL e.g. http://localhost:3000/verify?serial=XYZ or plain serial
    let serial = rawText.trim();
    try {
      const url = new URL(rawText);
      const urlSerial = url.searchParams.get('serial');
      if (urlSerial) {
        serial = urlSerial;
      } else {
        const parts = url.pathname.split('/');
        serial = parts[parts.length - 1];
      }
    } catch {
      // not a url, use direct string
    }

    if (scannerRef.current && scannerRef.current.isScanning) {
      scannerRef.current.stop().then(() => {
        onScan(serial);
      });
    } else {
      onScan(serial);
    }
  };

  const handleFileUpload = async (e) => {
    const file = e.target.files[0];
    if (!file || !scannerRef.current) return;

    try {
      const result = await scannerRef.current.scanFile(file, true);
      handleSuccess(result);
    } catch (err) {
      setError('Could not detect a valid QR code in the uploaded image.');
    }
  };

  return (
    <div className="glass-panel" style={{ padding: '1.5rem', margin: '1.5rem 0', position: 'relative' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
          <Camera size={20} color="var(--primary)" />
          <h3 style={{ fontSize: '1.1rem' }}>Live QR Camera Scanner</h3>
        </div>
        <button
          onClick={onClose}
          className="btn btn-ghost btn-sm"
          style={{ padding: '0.3rem', borderRadius: '50%' }}
        >
          <X size={18} />
        </button>
      </div>

      {error ? (
        <div style={{ padding: '1.5rem', textAlign: 'center', background: 'rgba(244, 63, 94, 0.1)', borderRadius: 'var(--radius-md)', border: '1px solid var(--border-subtle)' }}>
          <p style={{ color: 'var(--danger)', fontSize: '0.9rem', marginBottom: '1rem' }}>{error}</p>
          <input
            type="file"
            accept="image/*"
            ref={fileInputRef}
            onChange={handleFileUpload}
            style={{ display: 'none' }}
          />
          <button
            onClick={() => fileInputRef.current?.click()}
            className="btn btn-outline btn-sm"
          >
            <Upload size={16} /> Upload QR Image File Instead
          </button>
        </div>
      ) : (
        <div>
          <div
            id="qr-reader-container"
            style={{
              width: '100%',
              maxWidth: '380px',
              margin: '0 auto',
              borderRadius: 'var(--radius-md)',
              overflow: 'hidden',
              border: '2px solid var(--primary)',
            }}
          />
          <p style={{ textAlign: 'center', color: 'var(--text-muted)', fontSize: '0.85rem', marginTop: '0.75rem' }}>
            Point your camera at the QR code on the drug packaging.
          </p>
          <div style={{ textAlign: 'center', marginTop: '0.5rem' }}>
            <input
              type="file"
              accept="image/*"
              ref={fileInputRef}
              onChange={handleFileUpload}
              style={{ display: 'none' }}
            />
            <button
              onClick={() => fileInputRef.current?.click()}
              className="btn btn-ghost btn-sm"
              style={{ fontSize: '0.8rem' }}
            >
              <Upload size={14} /> Scan from Image File
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
